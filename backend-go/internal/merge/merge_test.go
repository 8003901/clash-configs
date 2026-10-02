package merge_test

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/8003901/clash-configs/backend-go/internal/merge"
)

const tmpl = `{
  "proxies": [],
  "proxy-groups": [
    {"name":"Main Node","type":"select","proxies":["LoadBalance"],"filter-key":"all"},
    {"name":"US","type":"url-test","filter-key":"US|美国"},
    {"name":"Auto","type":"url-test","filter-key":"all"}
  ],
  "rules": ["MATCH,Main Node","DOMAIN-SUFFIX,x.com,US","DOMAIN-SUFFIX,y.com,Gone"]
}`

const subA = `proxies:
  - name: "US-01"
    type: ss
    server: 1.1.1.1
    port: 443
  - name: "HK-01"
    type: ss
    server: 2.2.2.2
    port: 443
`

const subB = `proxies:
  - name: "US-02"
    type: vmess
    server: 3.3.3.3
    port: 8443
`

func TestMergeBasic(t *testing.T) {
	res, err := merge.Merge(tmpl, []merge.Source{
		{Content: subA, UserInfo: "upload=100; download=200; total=1000; expire=100"},
		{Content: subB, UserInfo: "upload=10; download=20; total=500; expire=200"},
	})
	if err != nil {
		t.Fatalf("merge: %v", err)
	}
	// proxies merged
	for _, want := range []string{"US-01", "HK-01", "US-02"} {
		if !strings.Contains(res.YAML, want) {
			t.Fatalf("missing proxy %q in output:\n%s", want, res.YAML)
		}
	}
	// filter-key removed
	if strings.Contains(res.YAML, "filter-key") {
		t.Fatalf("filter-key not removed:\n%s", res.YAML)
	}
	// userinfo aggregated
	if res.UserInfo == nil {
		t.Fatal("expected userinfo")
	}
	if res.UserInfo.Upload != 110 || res.UserInfo.Download != 220 || res.UserInfo.Total != 1500 || res.UserInfo.Expire != 100 {
		t.Fatalf("wrong aggregate: %+v", res.UserInfo)
	}
	// rule with unknown target replaced by Main Node
	if !strings.Contains(res.YAML, "DOMAIN-SUFFIX,y.com,Main Node") {
		t.Fatalf("rule not fixed:\n%s", res.YAML)
	}
	// MATCH rule kept
	if !strings.Contains(res.YAML, "MATCH,Main Node") {
		t.Fatalf("match rule lost:\n%s", res.YAML)
	}
}

func TestMergeFilterKeyMatchesSubset(t *testing.T) {
	res, err := merge.Merge(tmpl, []merge.Source{{Content: subA}})
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	if err := yaml.Unmarshal([]byte(res.YAML), &out); err != nil {
		t.Fatalf("parse output yaml: %v", err)
	}
	groups, _ := out["proxy-groups"].([]any)
	byName := map[string]map[string]any{}
	for _, g := range groups {
		gm, _ := g.(map[string]any)
		byName[gm["name"].(string)] = gm
	}
	us := stringSlice(byName["US"]["proxies"].([]any))
	if !contains(us, "US-01") || contains(us, "HK-01") {
		t.Fatalf("US group (filter US|美国) should contain US-01 but not HK-01: %v", us)
	}
	main := stringSlice(byName["Main Node"]["proxies"].([]any))
	if !contains(main, "US-01") || !contains(main, "HK-01") {
		t.Fatalf("Main Node group (filter all) should contain both: %v", main)
	}
}

func stringSlice(in []any) []string {
	out := make([]string, 0, len(in))
	for _, v := range in {
		if s, ok := v.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func contains(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}

func TestMergeSkipsEmptyAndOverQuota(t *testing.T) {
	// empty content skipped; over-quota (used>95%) skipped
	res, err := merge.Merge(tmpl, []merge.Source{
		{Content: "", UserInfo: "upload=1; download=1; total=10; expire=1"},
		{Content: subA, UserInfo: "upload=970; download=0; total=1000; expire=1"}, // 97% used
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(res.YAML, "US-01") {
		t.Fatalf("over-quota config should be skipped:\n%s", res.YAML)
	}
	if res.UserInfo != nil {
		t.Fatalf("all configs skipped, userinfo should be nil: %+v", res.UserInfo)
	}
}

func TestHeaderValue(t *testing.T) {
	d := merge.DataUsage{Upload: 1, Download: 2, Total: 3, Expire: 4}
	if got := d.HeaderValue(); got != "upload=1; download=2; total=3; expire=4" {
		t.Fatalf("got %q", got)
	}
}

// 两个机场的真实线路会撞名；合并后保留线路并生成唯一名称，避免 mihomo 拒绝配置。
const subDupA = `{"proxies":[{"name":"香港01","type":"ss","server":"a.example.com","port":443},{"name":"US-01","type":"ss","server":"1.1.1.1","port":443},{"name":"香港01 (2)","type":"ss","server":"alias.example.com","port":443}]}`

const subDupB = `{"proxies":[{"name":"香港01","type":"ss","server":"b.example.com","port":443},{"name":"US-02","type":"vmess","server":"3.3.3.3","port":8443},{"name":"香港01 (2)","type":"ss","server":"alias.example.com","port":443}]}`

func TestMergePreservesDuplicateProxyNames(t *testing.T) {
	res, err := merge.Merge(tmpl, []merge.Source{{Content: subDupA}, {Content: subDupB}})
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	if err := yaml.Unmarshal([]byte(res.YAML), &out); err != nil {
		t.Fatalf("parse output yaml: %v", err)
	}

	proxies, _ := out["proxies"].([]any)
	servers := map[string]string{}
	for _, p := range proxies {
		pm, _ := p.(map[string]any)
		n, _ := pm["name"].(string)
		servers[n], _ = pm["server"].(string)
	}
	for name, wantServer := range map[string]string{
		"香港01":     "a.example.com",
		"香港01 (2)": "alias.example.com",
		"香港01 (3)": "b.example.com",
		"US-01":    "1.1.1.1",
		"US-02":    "3.3.3.3",
	} {
		if got := servers[name]; got != wantServer {
			t.Fatalf("proxy %q server = %q, want %q; all servers: %v", name, got, wantServer, servers)
		}
	}
	if len(servers) != 5 {
		t.Fatalf("应保留全部 5 条线路，实际 %d 条: %v", len(servers), servers)
	}

	// 组的 proxies 列表里也不能出现重复名字
	groups, _ := out["proxy-groups"].([]any)
	for _, g := range groups {
		gm, _ := g.(map[string]any)
		gn, _ := gm["name"].(string)
		names := stringSlice(gm["proxies"].([]any))
		got := map[string]int{}
		for _, n := range names {
			got[n]++
		}
		for n, c := range got {
			if c > 1 {
				t.Fatalf("组 %q 的 proxies 里 %q 重复 %d 次: %v", gn, n, c, names)
			}
		}
	}
}

// 机场会把真实节点的定义复制几份、把 name 改成套餐信息（剩余流量/套餐到期/官网…），
// 在客户端里冒充成一个节点。这些“信息节点”不是线路，应当整条丢弃。
const subInfo = `proxies:
  - name: "剩余流量：1024 GB"
    type: vless
    server: jp.example.com
    port: 443
    uuid: 22962a38-b2ea-46b4-8d3a-65f7161c94c5
  - name: "套餐到期：长期有效"
    type: vless
    server: jp.example.com
    port: 443
    uuid: 22962a38-b2ea-46b4-8d3a-65f7161c94c5
  - name: "官网：example.com"
    type: vless
    server: jp.example.com
    port: 443
    uuid: 22962a38-b2ea-46b4-8d3a-65f7161c94c5
  - name: "订阅到期：2026-12-31"
    type: vless
    server: jp.example.com
    port: 443
    uuid: 22962a38-b2ea-46b4-8d3a-65f7161c94c5
  - name: "\U0001F1EF\U0001F1F5日本高速01|CTCU|0.5x"
    type: vless
    server: jp.example.com
    port: 443
    uuid: 22962a38-b2ea-46b4-8d3a-65f7161c94c5
`

func TestMergeDropsInfoNodes(t *testing.T) {
	res, err := merge.Merge(tmpl, []merge.Source{{Content: subInfo}})
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	if err := yaml.Unmarshal([]byte(res.YAML), &out); err != nil {
		t.Fatalf("parse output yaml: %v", err)
	}

	proxies, _ := out["proxies"].([]any)
	var names []string
	for _, p := range proxies {
		pm, _ := p.(map[string]any)
		n, _ := pm["name"].(string)
		names = append(names, n)
	}
	if len(names) != 1 {
		t.Fatalf("信息节点应全部丢弃，只剩 1 个真实节点，实际 %d 个: %v", len(names), names)
	}
	if names[0] != "🇯🇵日本高速01|CTCU|0.5x" {
		t.Fatalf("留下的应是真实节点，实际 %q", names[0])
	}

	// 组里不能再引用被丢掉的那几个信息节点
	dropped := []string{"剩余流量：1024 GB", "套餐到期：长期有效", "官网：example.com", "订阅到期：2026-12-31"}
	groups, _ := out["proxy-groups"].([]any)
	for _, g := range groups {
		gm, _ := g.(map[string]any)
		for _, n := range stringSlice(gm["proxies"].([]any)) {
			if contains(dropped, n) {
				t.Fatalf("组 %v 仍引用了已丢弃的信息节点 %q", gm["name"], n)
			}
		}
	}
}
