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
