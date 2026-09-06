package merge

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type Source struct {
	Content  string
	UserInfo string
}

type DataUsage struct {
	Upload   int64
	Download int64
	Total    int64
	Expire   int64
}

func (d DataUsage) HeaderValue() string {
	return fmt.Sprintf("upload=%d; download=%d; total=%d; expire=%d", d.Upload, d.Download, d.Total, d.Expire)
}

type Result struct {
	YAML     string
	UserInfo *DataUsage
}

const UsageLimitPercent = 95

var specialGroups = map[string]bool{
	"DIRECT": true, "REJECT": true, "REJECT-DROP": true, "REJECT-INT": true, "MATCH": true,
}

func Merge(template string, configs []Source) (*Result, error) {
	var tree map[string]any
	if err := json.Unmarshal([]byte(template), &tree); err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}

	var included []Source
	for _, c := range configs {
		if strings.TrimSpace(c.Content) == "" {
			continue
		}
		if overQuota(c.UserInfo) {
			continue
		}
		included = append(included, c)
	}

	proxies := getArray(tree, "proxies")
	for _, c := range included {
		var sub map[string]any
		if err := yaml.Unmarshal([]byte(c.Content), &sub); err != nil {
			continue
		}
		if sp, ok := sub["proxies"].([]any); ok {
			proxies = append(proxies, sp...)
		}
	}
	tree["proxies"] = proxies

	var proxyNames []string
	for _, p := range proxies {
		if m, ok := p.(map[string]any); ok {
			if n, ok := m["name"].(string); ok {
				proxyNames = append(proxyNames, n)
			}
		}
	}

	groups, _ := tree["proxy-groups"].([]any)
	for _, g := range groups {
		gm, ok := g.(map[string]any)
		if !ok {
			continue
		}
		filterKey, _ := gm["filter-key"].(string)
		switch {
		case filterKey == "all":
			addProxies(gm, toStrings(proxyNames))
		case filterKey != "":
			addProxies(gm, filterNames(proxies, strings.Split(filterKey, "|")))
		}
		delete(gm, "filter-key")
	}

	var kept []any
	for _, g := range groups {
		gm, ok := g.(map[string]any)
		if !ok {
			kept = append(kept, g)
			continue
		}
		ps, _ := gm["proxies"].([]any)
		if len(ps) == 0 {
			removeReferenceFromAll(groups, gm["name"])
			continue
		}
		kept = append(kept, g)
	}
	tree["proxy-groups"] = kept

	checkRules(tree)

	out, err := yaml.Marshal(tree)
	if err != nil {
		return nil, fmt.Errorf("marshal yaml: %w", err)
	}

	res := &Result{YAML: strings.TrimPrefix(string(out), "---\n")}
	res.UserInfo = aggregate(included)
	return res, nil
}

func getArray(m map[string]any, key string) []any {
	if a, ok := m[key].([]any); ok {
		return a
	}
	a := []any{}
	m[key] = a
	return a
}

func addProxies(group map[string]any, names []any) {
	if existing, ok := group["proxies"].([]any); ok {
		group["proxies"] = append(existing, names...)
	} else {
		group["proxies"] = names
	}
}

func toStrings(ss []string) []any {
	out := make([]any, len(ss))
	for i, s := range ss {
		out[i] = s
	}
	return out
}

func filterNames(proxies []any, keys []string) []any {
	var out []any
	for _, p := range proxies {
		m, ok := p.(map[string]any)
		if !ok {
			continue
		}
		name, _ := m["name"].(string)
		for _, k := range keys {
			if strings.Contains(name, k) {
				out = append(out, name)
				break
			}
		}
	}
	return out
}

func removeReferenceFromAll(groups []any, nameAny any) {
	name, _ := nameAny.(string)
	for _, other := range groups {
		om, ok := other.(map[string]any)
		if !ok {
			continue
		}
		ops, _ := om["proxies"].([]any)
		var filtered []any
		for _, p := range ops {
			if s, ok := p.(string); ok && s == name {
				continue
			}
			filtered = append(filtered, p)
		}
		om["proxies"] = filtered
	}
}

func checkRules(tree map[string]any) {
	groups, _ := tree["proxy-groups"].([]any)
	groupNames := map[string]bool{}
	for _, g := range groups {
		if gm, ok := g.(map[string]any); ok {
			if n, ok := gm["name"].(string); ok {
				groupNames[n] = true
			}
		}
	}

	rules, _ := tree["rules"].([]any)
	for i, r := range rules {
		rule, ok := r.(string)
		if !ok {
			continue
		}
		parts := strings.Split(rule, ",")
		targetIndex := 1
		if len(parts) >= 3 {
			targetIndex = 2
		}
		if len(parts) <= targetIndex {
			continue
		}
		target := strings.TrimSpace(parts[targetIndex])
		if target != "" && !specialGroups[target] && !groupNames[target] {
			parts[targetIndex] = "Main Node"
			rules[i] = strings.Join(parts, ",")
		}
	}
	tree["rules"] = rules
}

func overQuota(userInfo string) bool {
	if strings.TrimSpace(userInfo) == "" {
		return false
	}
	u := parseUserInfo(userInfo)
	if u.Total <= 0 {
		return false
	}
	used := u.Upload + u.Download
	return used*100 > u.Total*UsageLimitPercent
}

func parseUserInfo(info string) DataUsage {
	var u DataUsage
	for _, pair := range strings.Split(info, ";") {
		kv := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(kv) != 2 {
			continue
		}
		n, err := strconv.ParseInt(strings.TrimSpace(kv[1]), 10, 64)
		if err != nil {
			continue
		}
		switch strings.TrimSpace(kv[0]) {
		case "upload":
			u.Upload = n
		case "download":
			u.Download = n
		case "total":
			u.Total = n
		case "expire":
			u.Expire = n
		}
	}
	return u
}

func aggregate(configs []Source) *DataUsage {
	var agg *DataUsage
	for _, c := range configs {
		if strings.TrimSpace(c.UserInfo) == "" {
			continue
		}
		u := parseUserInfo(c.UserInfo)
		if agg == nil {
			agg = &u
			continue
		}
		agg.Upload += u.Upload
		agg.Download += u.Download
		agg.Total += u.Total
		if u.Expire < agg.Expire {
			agg.Expire = u.Expire
		}
	}
	return agg
}
