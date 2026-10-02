package main

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/8003901/clash-configs/backend-go/internal/merge"
)

func countProxies(label, content string) {
	var m map[string]any
	if err := yaml.Unmarshal([]byte(content), &m); err != nil {
		fmt.Printf("%-14s yaml error: %v\n", label, err)
		return
	}
	ps, _ := m["proxies"].([]any)
	seen := map[string]int{}
	for _, p := range ps {
		if pm, ok := p.(map[string]any); ok {
			n, _ := pm["name"].(string)
			seen[n]++
		}
	}
	dups := 0
	for _, c := range seen {
		if c > 1 {
			dups++
		}
	}
	fmt.Printf("%-14s proxies=%-4d unique=%-4d dupNames=%d\n", label, len(ps), len(seen), dups)
}

func main() {
	dir := os.Args[1]
	tmplPath := os.Args[2]
	names := os.Args[3:]

	tmpl, err := os.ReadFile(tmplPath)
	if err != nil {
		panic(err)
	}

	var srcs []merge.Source
	fmt.Println("=== inputs ===")
	for _, n := range names {
		b, err := os.ReadFile(dir + "/" + n + ".yaml")
		if err != nil {
			panic(err)
		}
		countProxies(n, string(b))
		srcs = append(srcs, merge.Source{Content: string(b)})
	}

	res, err := merge.Merge(string(tmpl), srcs)
	if err != nil {
		panic(err)
	}

	fmt.Println("=== merged ===")
	countProxies("MERGED", res.YAML)

	var out map[string]any
	_ = yaml.Unmarshal([]byte(res.YAML), &out)
	ps, _ := out["proxies"].([]any)
	fmt.Printf("\nmerged proxy names (%d):\n", len(ps))
	for i, p := range ps {
		pm, _ := p.(map[string]any)
		n, _ := pm["name"].(string)
		srv, _ := pm["server"].(string)
		typ, _ := pm["type"].(string)
		fmt.Printf("  %3d  %-40s %-8s %s\n", i+1, n, typ, srv)
	}

	groups, _ := out["proxy-groups"].([]any)
	fmt.Printf("\nproxy-groups kept: %d\n", len(groups))
	for _, g := range groups {
		gm, _ := g.(map[string]any)
		gn, _ := gm["name"].(string)
		gps, _ := gm["proxies"].([]any)
		fmt.Printf("  %-18s type=%-12s n=%d\n", gn, gm["type"], len(gps))
	}
	_ = strings.TrimSpace
}
