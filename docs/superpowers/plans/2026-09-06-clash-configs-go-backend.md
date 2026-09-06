# clash-configs Go 后端重构 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在 `backend-go/` 用 Go（Gin + GORM + 纯 Go SQLite）忠实 1:1 重写现有 Kotlin/Spring 后端，现有 Vue 前端零改动即可对接。

**Architecture:** 分层：`model`（GORM 模型）→ `store`（数据访问）→ `service`（业务）→ `web`（Gin handler + 鉴权 + session）→ `scheduler`（定时刷新）；`merge` 是纯函数合并引擎。前端构建产物经 `go:embed` 打进单一静态二进制。

**Tech Stack:** Go 1.24、Gin、GORM + `github.com/glebarez/sqlite`（纯 Go SQLite）、`golang.org/x/crypto/bcrypt`、`gopkg.in/yaml.v3`、`github.com/robfig/cron/v3`、`github.com/google/uuid`。

**Spec:** [2026-09-06-clash-configs-go-backend-design.md](../specs/2026-09-06-clash-configs-go-backend-design.md)

## Global Constraints

- Go 模块路径：`github.com/8003901/clash-configs/backend-go`；Go 版本 `1.26`（`golang.org/x/crypto` 最新版要求 ≥1.26，构建时自动解析到 1.26）。
- 编译目标 `CGO_ENABLED=0`（纯静态二进制，无 C 依赖）。
- JSON 字段名一律 camelCase：`updateSchedule`、`subscriptionUserinfo`、`createdAt`、`updatedAt`、`configs`。
- 时间序列化为 RFC3339 字符串（`time.Time` 的默认 JSON 编码），字段非空（GORM 自动维护 `CreatedAt`/`UpdatedAt`）。
- 数据库默认路径 `data/demo.db`；`SERVER_PORT` 环境变量控制监听端口（默认 8080）。
- 每次 `git commit` 消息以 `Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>` 结尾。

### 有意偏离 1:1 的行为（已与用户确认或记录在案）

1. **已确认修复**：`GET /clash_configs` 返回**全部**配置（含 `enabled=false`），不再用 `enabled=true` 过滤。
2. **记录在案（忠实复刻）**：`PUT /clash_configs/{id}` 更新时**不更新** `updateSchedule` 字段（原实现如此，疑似 bug，执行时保持复刻，见 Task 3 注释）。
3. **轻微偏离**：`PUT /clash_configs/{id}` 更新不存在的 id → 返回 404（原实现返回 200 + 未落库对象，属缺陷）。
4. **已接受取舍**：YAML 输出键按字母排序（`yaml.v3` 对 `map` 的默认行为），功能等价、非字节一致。

---

## Task 1: 项目脚手架 + GORM 模型 + store 数据层

**Files:**
- Create: `backend-go/go.mod`（由 `go mod init` 生成）
- Create: `backend-go/internal/model/model.go`
- Create: `backend-go/internal/store/store.go`
- Test: `backend-go/internal/store/store_test.go`

**Interfaces:**
- Consumes: 无（首个任务）。
- Produces:
  - `model.ClashConfig`、`model.ClashConfigsMerge`、`model.User` 结构体（字段见下）。
  - `store.New(db *gorm.DB) *Store`、`(*Store).Migrate() error`
  - `(*Store).SaveClashConfig(c *model.ClashConfig) error`
  - `(*Store).CreateClashConfig(c *model.ClashConfig) error`
  - `(*Store).FindClashConfig(id string) (*model.ClashConfig, error)`
  - `(*Store).ListClashConfigs() ([]model.ClashConfig, error)`
  - `(*Store).FindClashConfigsByIDs(ids []string) ([]model.ClashConfig, error)`
  - `(*Store).DeleteClashConfig(id string) error`
  - `(*Store).ListClashConfigsBySchedule(schedule string) ([]model.ClashConfig, error)`
  - `(*Store).CreateMerge(m *model.ClashConfigsMerge) error`
  - `(*Store).SaveMerge(m *model.ClashConfigsMerge) error`
  - `(*Store).FindMerge(id string) (*model.ClashConfigsMerge, error)`
  - `(*Store).ListMerges() ([]model.ClashConfigsMerge, error)`
  - `(*Store).FindMergeByToken(token string) (*model.ClashConfigsMerge, error)`
  - `(*Store).FindUserByUsername(username string) (*model.User, error)`
  - `(*Store).SaveUser(u *model.User) error`

- [ ] **Step 1: 初始化模块并安装依赖**

```bash
mkdir -p backend-go/internal/{model,store,service,merge,web,scheduler,config} backend-go/cmd/{server,migrate}
cd backend-go
go mod init github.com/8003901/clash-configs/backend-go
go get gorm.io/gorm github.com/glebarez/sqlite github.com/google/uuid \
  golang.org/x/crypto/bcrypt gopkg.in/yaml.v3 github.com/robfig/cron/v3 \
  github.com/gin-gonic/gin
```

- [ ] **Step 2: 写失败测试**

创建 `backend-go/internal/store/store_test.go`：

```go
package store_test

import (
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/8003901/clash-configs/backend-go/internal/model"
	"github.com/8003901/clash-configs/backend-go/internal/store"
)

func newTestStore(t *testing.T) *store.Store {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "test.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	s := store.New(db)
	if err := s.Migrate(); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return s
}

func TestClashConfigCRUD(t *testing.T) {
	s := newTestStore(t)
	c := &model.ClashConfig{
		ID: "id-1", URL: "https://example.com/sub", Name: "foo",
		Enabled: true, UpdateSchedule: "DAY", Content: "proxies: []",
		SubscriptionUserinfo: "upload=1; download=2; total=3; expire=4",
	}
	if err := s.CreateClashConfig(c); err != nil {
		t.Fatalf("create: %v", err)
	}
	if c.CreatedAt.IsZero() || c.UpdatedAt.IsZero() {
		t.Fatalf("auto timestamps not set: created=%v updated=%v", c.CreatedAt, c.UpdatedAt)
	}
	got, err := s.FindClashConfig("id-1")
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if got.Name != "foo" || got.SubscriptionUserinfo != "upload=1; download=2; total=3; expire=4" {
		t.Fatalf("unexpected got: %+v", got)
	}
	list, err := s.ListClashConfigs()
	if err != nil || len(list) != 1 {
		t.Fatalf("list: n=%d err=%v", len(list), err)
	}
	if err := s.DeleteClashConfig("id-1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := s.FindClashConfig("id-1"); err != gorm.ErrRecordNotFound {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}

func TestMergeAssociation(t *testing.T) {
	s := newTestStore(t)
	c1 := &model.ClashConfig{ID: "c1", Name: "a", Content: "x"}
	c2 := &model.ClashConfig{ID: "c2", Name: "b", Content: "y"}
	_ = s.CreateClashConfig(c1)
	_ = s.CreateClashConfig(c2)

	m := &model.ClashConfigsMerge{ID: "m1", Name: "merge", Token: "t1", Config: "{}", Configs: []model.ClashConfig{*c1, *c2}}
	if err := s.CreateMerge(m); err != nil {
		t.Fatalf("create merge: %v", err)
	}
	got, err := s.FindMerge("m1")
	if err != nil {
		t.Fatalf("find merge: %v", err)
	}
	if len(got.Configs) != 2 {
		t.Fatalf("expected 2 configs, got %d", len(got.Configs))
	}
	byToken, err := s.FindMergeByToken("t1")
	if err != nil || byToken.Name != "merge" {
		t.Fatalf("find by token: %+v err=%v", byToken, err)
	}

	// update: replace association to just c1
	m.Configs = []model.ClashConfig{*c1}
	if err := s.SaveMerge(m); err != nil {
		t.Fatalf("save merge: %v", err)
	}
	got, _ = s.FindMerge("m1")
	if len(got.Configs) != 1 || got.Configs[0].ID != "c1" {
		t.Fatalf("association not replaced: %+v", got.Configs)
	}
}

func TestUserCRUD(t *testing.T) {
	s := newTestStore(t)
	u := &model.User{ID: "u1", Username: "admin", Password: "hash"}
	if err := s.SaveUser(u); err != nil {
		t.Fatalf("save user: %v", err)
	}
	got, err := s.FindUserByUsername("admin")
	if err != nil || got.Password != "hash" {
		t.Fatalf("find user: %+v err=%v", got, err)
	}
	if _, err := s.FindUserByUsername("nobody"); err != gorm.ErrRecordNotFound {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}
```

- [ ] **Step 3: 运行测试确认失败**

Run: `cd backend-go && go test ./internal/store/ -v`
Expected: FAIL（`package .../store` 不存在 / 未定义 `store.New`、`model.ClashConfig`）。

- [ ] **Step 4: 写模型**

创建 `backend-go/internal/model/model.go`：

```go
package model

import "time"

type ClashConfig struct {
	ID                   string    `gorm:"primaryKey;size:36" json:"id"`
	URL                  string    `json:"url"`
	Name                 string    `json:"name"`
	Enabled              bool      `json:"enabled"`
	UpdateSchedule       string    `json:"updateSchedule"`
	Content              string    `json:"content"`
	CreatedAt            time.Time `json:"createdAt"`
	UpdatedAt            time.Time `json:"updatedAt"`
	SubscriptionUserinfo string    `json:"subscriptionUserinfo"`
}

type ClashConfigsMerge struct {
	ID        string        `gorm:"primaryKey;size:36" json:"id"`
	Name      string        `json:"name"`
	Token     string        `json:"token"`
	Config    string        `json:"config"`
	CreatedAt time.Time     `json:"createdAt"`
	UpdatedAt time.Time     `json:"updatedAt"`
	Configs   []ClashConfig `gorm:"many2many:clash_configs_merge_config" json:"configs"`
}

type User struct {
	ID       string `gorm:"primaryKey;size:36" json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}
```

- [ ] **Step 5: 写 store**

创建 `backend-go/internal/store/store.go`：

```go
package store

import (
	"gorm.io/gorm"

	"github.com/8003901/clash-configs/backend-go/internal/model"
)

type Store struct {
	DB *gorm.DB
}

func New(db *gorm.DB) *Store { return &Store{DB: db} }

func (s *Store) Migrate() error {
	return s.DB.AutoMigrate(&model.ClashConfig{}, &model.ClashConfigsMerge{}, &model.User{})
}

// --- ClashConfig ---

func (s *Store) CreateClashConfig(c *model.ClashConfig) error { return s.DB.Create(c).Error }
func (s *Store) SaveClashConfig(c *model.ClashConfig) error   { return s.DB.Save(c).Error }

func (s *Store) FindClashConfig(id string) (*model.ClashConfig, error) {
	var c model.ClashConfig
	if err := s.DB.First(&c, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Store) ListClashConfigs() ([]model.ClashConfig, error) {
	var cs []model.ClashConfig
	err := s.DB.Order("created_at ASC").Find(&cs).Error
	return cs, err
}

func (s *Store) FindClashConfigsByIDs(ids []string) ([]model.ClashConfig, error) {
	var cs []model.ClashConfig
	if len(ids) == 0 {
		return cs, nil
	}
	err := s.DB.Where("id IN ?", ids).Find(&cs).Error
	return cs, err
}

func (s *Store) DeleteClashConfig(id string) error {
	return s.DB.Delete(&model.ClashConfig{}, "id = ?", id).Error
}

func (s *Store) ListClashConfigsBySchedule(schedule string) ([]model.ClashConfig, error) {
	var cs []model.ClashConfig
	err := s.DB.Where("update_schedule = ?", schedule).Find(&cs).Error
	return cs, err
}

// --- ClashConfigsMerge ---

// CreateMerge 新建 merge 并写入多对多关联（先插入主记录，再替换 join 表）。
func (s *Store) CreateMerge(m *model.ClashConfigsMerge) error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Configs").Create(m).Error; err != nil {
			return err
		}
		return tx.Model(m).Association("Configs").Replace(m.Configs)
	})
}

// SaveMerge 更新 merge（主键已存在）并整体替换多对多关联。
func (s *Store) SaveMerge(m *model.ClashConfigsMerge) error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Configs").Save(m).Error; err != nil {
			return err
		}
		return tx.Model(m).Association("Configs").Replace(m.Configs)
	})
}

func (s *Store) FindMerge(id string) (*model.ClashConfigsMerge, error) {
	var m model.ClashConfigsMerge
	if err := s.DB.Preload("Configs").First(&m, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *Store) ListMerges() ([]model.ClashConfigsMerge, error) {
	var ms []model.ClashConfigsMerge
	err := s.DB.Preload("Configs").Find(&ms).Error
	return ms, err
}

func (s *Store) FindMergeByToken(token string) (*model.ClashConfigsMerge, error) {
	var m model.ClashConfigsMerge
	if err := s.DB.Preload("Configs").First(&m, "token = ?", token).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

// --- User ---

func (s *Store) FindUserByUsername(username string) (*model.User, error) {
	var u model.User
	if err := s.DB.First(&u, "username = ?", username).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *Store) SaveUser(u *model.User) error { return s.DB.Save(u).Error }
```

- [ ] **Step 6: 运行测试确认通过**

Run: `cd backend-go && go test ./internal/store/ -v`
Expected: PASS（3 个测试）。

- [ ] **Step 7: 提交**

```bash
git add backend-go/go.mod backend-go/go.sum backend-go/internal/model backend-go/internal/store
git commit -m "feat(backend-go): add GORM models and store data layer

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 2: 合并引擎（纯函数核心）

**Files:**
- Create: `backend-go/internal/merge/merge.go`
- Test: `backend-go/internal/merge/merge_test.go`

**Interfaces:**
- Consumes: 无（独立纯函数，不依赖其它任务）。
- Produces:
  - `merge.Source{ Content, UserInfo string }`
  - `merge.DataUsage{ Upload, Download, Total, Expire int64 }`，方法 `HeaderValue() string`
  - `merge.Result{ YAML string; UserInfo *DataUsage }`
  - `merge.Merge(template string, configs []Source) (*Result, error)`
  - 常量 `merge.UsageLimitPercent = 95`

- [ ] **Step 1: 写失败测试**

创建 `backend-go/internal/merge/merge_test.go`：

```go
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
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd backend-go && go test ./internal/merge/ -v`
Expected: FAIL（`package .../merge` 不存在）。

- [ ] **Step 3: 实现合并引擎**

创建 `backend-go/internal/merge/merge.go`：

```go
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
```

- [ ] **Step 4: 运行测试确认通过**

Run: `cd backend-go && go test ./internal/merge/ -v`
Expected: PASS。

- [ ] **Step 5: 提交**

```bash
git add backend-go/internal/merge
git commit -m "feat(backend-go): add merge engine

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 3: ClashConfig 服务（远程拉取 + CRUD）

**Files:**
- Create: `backend-go/internal/service/clash_config.go`
- Test: `backend-go/internal/service/clash_config_test.go`

**Interfaces:**
- Consumes: `store.Store`（Task 1）、`merge`（本任务不用，仅供后续）、`model`（Task 1）。
- Produces:
  - `service.NewClashConfigService(s *store.Store) *ClashConfigService`
  - `(*ClashConfigService).Save(cc *model.ClashConfig) (*model.ClashConfig, error)`
  - `(*ClashConfigService).Detail(id string, renew bool) (*model.ClashConfig, error)`
  - `(*ClashConfigService).All() ([]model.ClashConfig, error)`
  - `(*ClashConfigService).Delete(id string) error`
  - `(*ClashConfigService).Renew(cc *model.ClashConfig) error`
  - 错误哨兵 `service.ErrNotFound = errors.New("not found")`

- [ ] **Step 1: 写失败测试**（用 `httptest` 模拟远端订阅）

创建 `backend-go/internal/service/clash_config_test.go`：

```go
package service_test

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/8003901/clash-configs/backend-go/internal/model"
	"github.com/8003901/clash-configs/backend-go/internal/service"
	"github.com/8003901/clash-configs/backend-go/internal/store"
)

func newService(t *testing.T) (*service.ClashConfigService, *httptest.Server) {
	t.Helper()
	db, _ := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "t.db")), &gorm.Config{})
	s := store.New(db)
	_ = s.Migrate()
	svc := service.NewClashConfigService(s)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != "clash-verge/*" {
			t.Errorf("missing UA, got %q", r.Header.Get("User-Agent"))
		}
		w.Header().Set("subscription-userinfo", "upload=10; download=20; total=100; expire=99")
		_, _ = w.Write([]byte("proxies:\n  - name: n1\n"))
	}))
	return svc, ts
}

func TestSaveCreateAndList(t *testing.T) {
	svc, ts := newService(t)
	defer ts.Close()

	cc := &model.ClashConfig{URL: ts.URL, Name: "sub", Enabled: true, UpdateSchedule: "DAY"}
	got, err := svc.Save(cc)
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	if got.ID == "" || got.Content == "" || got.SubscriptionUserinfo != "upload=10; download=20; total=100; expire=99" {
		t.Fatalf("bad result: %+v", got)
	}

	list, err := svc.All()
	if err != nil || len(list) != 1 {
		t.Fatalf("all: n=%d err=%v", len(list), err)
	}
}

func TestAllIncludesDisabled(t *testing.T) {
	svc, ts := newService(t)
	defer ts.Close()
	// 修复点：停用配置也必须出现在列表中
	disabled := &model.ClashConfig{URL: ts.URL, Name: "off", Enabled: false, UpdateSchedule: "DAY"}
	if _, err := svc.Save(disabled); err != nil {
		t.Fatal(err)
	}
	list, _ := svc.All()
	if len(list) != 1 || list[0].Enabled {
		t.Fatalf("disabled config should be listed: %+v", list)
	}
}

func TestDetailRenewAndNotFound(t *testing.T) {
	svc, ts := newService(t)
	defer ts.Close()
	cc, _ := svc.Save(&model.ClashConfig{URL: ts.URL, Name: "x", Enabled: true, UpdateSchedule: "DAY"})

	d, err := svc.Detail(cc.ID, false)
	if err != nil || d.Name != "x" {
		t.Fatalf("detail: %+v err=%v", d, err)
	}
	if _, err := svc.Detail("missing", false); err != service.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if _, err := svc.Detail(cc.ID, true); err != nil {
		t.Fatalf("detail renew: %v", err)
	}
}

func TestUpdateNonExistentReturnsNotFound(t *testing.T) {
	svc, ts := newService(t)
	defer ts.Close()
	_, err := svc.Save(&model.ClashConfig{ID: "nope", URL: ts.URL, Name: "x"})
	if err != service.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd backend-go && go test ./internal/service/ -run TestSaveCreateAndList -v`
Expected: FAIL（`service` 包未定义）。

- [ ] **Step 3: 实现服务**

创建 `backend-go/internal/service/clash_config.go`：

```go
package service

import (
	"errors"
	"io"
	"net/http"

	"gorm.io/gorm"

	"github.com/8003901/clash-configs/backend-go/internal/model"
	"github.com/8003901/clash-configs/backend-go/internal/store"
	"github.com/google/uuid"
)

var ErrNotFound = errors.New("not found")

type ClashConfigService struct {
	store  *store.Store
	client *http.Client
}

func NewClashConfigService(s *store.Store) *ClashConfigService {
	return &ClashConfigService{store: s, client: &http.Client{}}
}

// Save 拉取远端订阅后创建或更新。
// 更新路径忠实复刻原实现：只更新 name/enabled/url/content/subscriptionUserinfo，
// 不更新 updateSchedule（原实现即如此）。
func (s *ClashConfigService) Save(cc *model.ClashConfig) (*model.ClashConfig, error) {
	if err := s.fetchRemote(cc); err != nil {
		return nil, err
	}
	if cc.ID != "" {
		existing, err := s.store.FindClashConfig(cc.ID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrNotFound
			}
			return nil, err
		}
		if cc.Name != "" {
			existing.Name = cc.Name
		}
		existing.Enabled = cc.Enabled
		existing.URL = cc.URL
		existing.SubscriptionUserinfo = cc.SubscriptionUserinfo
		existing.Content = cc.Content
		if err := s.store.SaveClashConfig(existing); err != nil {
			return nil, err
		}
		return existing, nil
	}
	cc.ID = uuid.NewString()
	if err := s.store.CreateClashConfig(cc); err != nil {
		return nil, err
	}
	return cc, nil
}

func (s *ClashConfigService) Detail(id string, renew bool) (*model.ClashConfig, error) {
	cc, err := s.store.FindClashConfig(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if renew {
		return cc, s.Renew(cc)
	}
	return cc, nil
}

func (s *ClashConfigService) All() ([]model.ClashConfig, error) {
	return s.store.ListClashConfigs()
}

func (s *ClashConfigService) Delete(id string) error {
	return s.store.DeleteClashConfig(id)
}

func (s *ClashConfigService) Renew(cc *model.ClashConfig) error {
	if err := s.fetchRemote(cc); err != nil {
		return err
	}
	return s.store.SaveClashConfig(cc)
}

func (s *ClashConfigService) fetchRemote(cc *model.ClashConfig) error {
	req, err := http.NewRequest(http.MethodGet, cc.URL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "clash-verge/*")
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	cc.Content = string(body)
	cc.SubscriptionUserinfo = resp.Header.Get("subscription-userinfo")
	return nil
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `cd backend-go && go test ./internal/service/ -run 'TestSaveCreateAndList|TestAllIncludesDisabled|TestDetailRenewAndNotFound|TestUpdateNonExistentReturnsNotFound' -v`
Expected: PASS。

- [ ] **Step 5: 提交**

```bash
git add backend-go/internal/service
git commit -m "feat(backend-go): add ClashConfig service with remote fetch

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 4: Merge 服务（模板加载 + CRUD）

**Files:**
- Create: `backend-go/internal/service/merge_service.go`
- Test: `backend-go/internal/service/merge_service_test.go`

**Interfaces:**
- Consumes: `store.Store`（Task 1）、`model`（Task 1）、`service.ErrNotFound`（Task 3）。
- Produces:
  - `service.NewMergeService(s *store.Store, template string) *MergeService`
  - `(*MergeService).Create(name string, configIDs []string) (*model.ClashConfigsMerge, error)`
  - `(*MergeService).Save(m *model.ClashConfigsMerge) (*model.ClashConfigsMerge, error)`
  - `(*MergeService).Detail(id string) (*model.ClashConfigsMerge, error)`
  - `(*MergeService).List() ([]model.ClashConfigsMerge, error)`
  - `(*MergeService).RefreshToken(id string) (*model.ClashConfigsMerge, error)`
  - `(*MergeService).FindByToken(token string) (*model.ClashConfigsMerge, error)`

- [ ] **Step 1: 写失败测试**

创建 `backend-go/internal/service/merge_service_test.go`：

```go
package service_test

import (
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/8003901/clash-configs/backend-go/internal/model"
	"github.com/8003901/clash-configs/backend-go/internal/service"
	"github.com/8003901/clash-configs/backend-go/internal/store"
)

const tmpl = `{"proxies":[],"proxy-groups":[],"rules":[]}`

func newMergeService(t *testing.T) (*service.MergeService, *store.Store) {
	t.Helper()
	db, _ := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "t.db")), &gorm.Config{})
	s := store.New(db)
	_ = s.Migrate()
	_ = s.CreateClashConfig(&model.ClashConfig{ID: "c1", Name: "a"})
	_ = s.CreateClashConfig(&model.ClashConfig{ID: "c2", Name: "b"})
	return service.NewMergeService(s, tmpl), s
}

func TestCreateAndDetail(t *testing.T) {
	svc, _ := newMergeService(t)
	m, err := svc.Create("merged", []string{"c1", "c2"})
	if err != nil {
		t.Fatal(err)
	}
	if m.ID == "" || m.Token == "" || m.Config != tmpl || len(m.Configs) != 2 {
		t.Fatalf("bad create: %+v", m)
	}
	got, err := svc.FindByToken(m.Token)
	if err != nil || got.Name != "merged" {
		t.Fatalf("find by token: %+v err=%v", got, err)
	}
	d, err := svc.Detail(m.ID)
	if err != nil || len(d.Configs) != 2 {
		t.Fatalf("detail: %+v err=%v", d, err)
	}
}

func TestSavePreservesTokenAndRefreshes(t *testing.T) {
	svc, _ := newMergeService(t)
	m, _ := svc.Create("merged", []string{"c1"})
	oldToken := m.Token
	oldCreated := m.CreatedAt

	m.Name = "renamed"
	m.Config = `{"x":1}`
	m.Configs = []model.ClashConfig{{ID: "c2"}}
	saved, err := svc.Save(m)
	if err != nil {
		t.Fatal(err)
	}
	if saved.Token != oldToken {
		t.Fatalf("token changed: %q != %q", saved.Token, oldToken)
	}
	if !saved.CreatedAt.Equal(oldCreated) {
		t.Fatalf("createdAt changed: %v != %v", saved.CreatedAt, oldCreated)
	}
	got, _ := svc.Detail(m.ID)
	if got.Name != "renamed" || got.Config != `{"x":1}` || len(got.Configs) != 1 || got.Configs[0].ID != "c2" {
		t.Fatalf("bad save: %+v", got)
	}
}

func TestRefreshTokenAndNotFound(t *testing.T) {
	svc, _ := newMergeService(t)
	m, _ := svc.Create("m", nil)
	old := m.Token
	refreshed, err := svc.RefreshToken(m.ID)
	if err != nil || refreshed.Token == old {
		t.Fatalf("refresh: %+v err=%v", refreshed, err)
	}
	if _, err := svc.Detail("nope"); err != service.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if _, err := svc.RefreshToken("nope"); err != service.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd backend-go && go test ./internal/service/ -run TestCreateAndDetail -v`
Expected: FAIL（`MergeService` 未定义）。

- [ ] **Step 3: 实现服务**

创建 `backend-go/internal/service/merge_service.go`：

```go
package service

import (
	"errors"

	"gorm.io/gorm"

	"github.com/8003901/clash-configs/backend-go/internal/model"
	"github.com/8003901/clash-configs/backend-go/internal/store"
	"github.com/google/uuid"
)

type MergeService struct {
	store    *store.Store
	template string
}

func NewMergeService(s *store.Store, template string) *MergeService {
	return &MergeService{store: s, template: template}
}

func (s *MergeService) Create(name string, configIDs []string) (*model.ClashConfigsMerge, error) {
	m := &model.ClashConfigsMerge{
		ID:     uuid.NewString(),
		Name:   name,
		Token:  uuid.NewString(),
		Config: s.template,
	}
	if len(configIDs) > 0 {
		configs, err := s.store.FindClashConfigsByIDs(configIDs)
		if err != nil {
			return nil, err
		}
		m.Configs = configs
	}
	if err := s.store.CreateMerge(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *MergeService) Save(m *model.ClashConfigsMerge) (*model.ClashConfigsMerge, error) {
	if len(m.Configs) > 0 {
		ids := make([]string, 0, len(m.Configs))
		for _, c := range m.Configs {
			if c.ID != "" {
				ids = append(ids, c.ID)
			}
		}
		if len(ids) > 0 {
			configs, err := s.store.FindClashConfigsByIDs(ids)
			if err != nil {
				return nil, err
			}
			m.Configs = configs
		} else {
			m.Configs = nil
		}
	}
	if m.ID != "" {
		existing, err := s.store.FindMerge(m.ID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrNotFound
			}
			return nil, err
		}
		m.Token = existing.Token
		m.CreatedAt = existing.CreatedAt
	} else {
		m.ID = uuid.NewString()
		m.Token = uuid.NewString()
	}
	if err := s.store.SaveMerge(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *MergeService) Detail(id string) (*model.ClashConfigsMerge, error) {
	m, err := s.store.FindMerge(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return m, nil
}

func (s *MergeService) List() ([]model.ClashConfigsMerge, error) {
	return s.store.ListMerges()
}

func (s *MergeService) RefreshToken(id string) (*model.ClashConfigsMerge, error) {
	m, err := s.store.FindMerge(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	m.Token = uuid.NewString()
	if err := s.store.SaveMerge(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *MergeService) FindByToken(token string) (*model.ClashConfigsMerge, error) {
	m, err := s.store.FindMergeByToken(token)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return m, nil
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `cd backend-go && go test ./internal/service/ -run 'TestCreateAndDetail|TestSavePreservesTokenAndRefreshes|TestRefreshTokenAndNotFound' -v`
Expected: PASS。

- [ ] **Step 5: 提交**

```bash
git add backend-go/internal/service
git commit -m "feat(backend-go): add merge service

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 5: User 服务 + session + 鉴权中间件

**Files:**
- Create: `backend-go/internal/service/user.go`
- Create: `backend-go/internal/web/session.go`
- Create: `backend-go/internal/web/auth.go`
- Test: `backend-go/internal/service/user_test.go`

**Interfaces:**
- Consumes: `store.Store`（Task 1）、`model`（Task 1）。
- Produces:
  - `service.NewUserService(s *store.Store) *UserService`
  - `(*UserService).EnsureAdmin(pwdInit bool) error`
  - `(*UserService).ChangePwd(username, newPwd, oldPwd string) error`
  - `(*UserService).Authenticate(username, password string) bool`
  - `service.ErrBadOldPassword = errors.New("旧密码错误")`
  - `web.NewSessionStore() *SessionStore`、`(*SessionStore).Create(username string) string`、`(*SessionStore).Get(id string) (string, bool)`、`(*SessionStore).Delete(id string)`

- [ ] **Step 1: 写失败测试**

创建 `backend-go/internal/service/user_test.go`：

```go
package service_test

import (
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/8003901/clash-configs/backend-go/internal/service"
	"github.com/8003901/clash-configs/backend-go/internal/store"
)

func newUserService(t *testing.T) *service.UserService {
	t.Helper()
	db, _ := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "t.db")), &gorm.Config{})
	s := store.New(db)
	_ = s.Migrate()
	return service.NewUserService(s)
}

func TestEnsureAdminCreatesDefault(t *testing.T) {
	u := newUserService(t)
	if err := u.EnsureAdmin(false); err != nil {
		t.Fatal(err)
	}
	if !u.Authenticate("admin", "password") {
		t.Fatal("default admin should authenticate with password")
	}
	if u.Authenticate("admin", "wrong") {
		t.Fatal("wrong password should not authenticate")
	}
}

func TestEnsureAdminIdempotentAndReset(t *testing.T) {
	u := newUserService(t)
	_ = u.EnsureAdmin(false)
	_ = u.ChangePwd("admin", "newpass", "password")
	// pwdInit=true 重置为默认密码
	if err := u.EnsureAdmin(true); err != nil {
		t.Fatal(err)
	}
	if !u.Authenticate("admin", "password") {
		t.Fatal("pwdInit should reset to default password")
	}
}

func TestChangePwdRejectsWrongOld(t *testing.T) {
	u := newUserService(t)
	_ = u.EnsureAdmin(false)
	err := u.ChangePwd("admin", "newpass", "wrongold")
	if err != service.ErrBadOldPassword {
		t.Fatalf("expected ErrBadOldPassword, got %v", err)
	}
	if !u.Authenticate("admin", "password") {
		t.Fatal("password should be unchanged after failed change")
	}
	// correct change
	if err := u.ChangePwd("admin", "newpass", "password"); err != nil {
		t.Fatal(err)
	}
	if !u.Authenticate("admin", "newpass") {
		t.Fatal("new password should authenticate")
	}
	// 迁移来的 bcrypt 哈希可直接校验
	hash, _ := bcrypt.GenerateFromPassword([]byte("legacy"), bcrypt.DefaultCost)
	if bcrypt.CompareHashAndPassword(hash, []byte("legacy")) != nil {
		t.Fatal("bcrypt roundtrip sanity check failed")
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd backend-go && go test ./internal/service/ -run TestEnsureAdminCreatesDefault -v`
Expected: FAIL（`UserService` 未定义）。

- [ ] **Step 3: 实现 User 服务**

创建 `backend-go/internal/service/user.go`：

```go
package service

import (
	"errors"

	"golang.org/x/crypto/bcrypt"

	"github.com/8003901/clash-configs/backend-go/internal/model"
	"github.com/8003901/clash-configs/backend-go/internal/store"
	"github.com/google/uuid"
)

var ErrBadOldPassword = errors.New("旧密码错误")

type UserService struct {
	store *store.Store
}

func NewUserService(s *store.Store) *UserService { return &UserService{store: s} }

// EnsureAdmin 启动时确保 admin 存在；pwdInit=true 时重置为默认密码。
func (s *UserService) EnsureAdmin(pwdInit bool) error {
	u, err := s.store.FindUserByUsername("admin")
	if err == nil && !pwdInit {
		return nil
	}
	if u == nil {
		u = &model.User{ID: uuid.NewString(), Username: "admin"}
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return s.store.SaveUser(u)
}

func (s *UserService) Authenticate(username, password string) bool {
	u, err := s.store.FindUserByUsername(username)
	if err != nil {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)) == nil
}

func (s *UserService) ChangePwd(username, newPwd, oldPwd string) error {
	u, err := s.store.FindUserByUsername(username)
	if err != nil {
		return ErrBadOldPassword
	}
	if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(oldPwd)) != nil {
		return ErrBadOldPassword
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPwd), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return s.store.SaveUser(u)
}
```

- [ ] **Step 4: 写 session 与鉴权中间件**

创建 `backend-go/internal/web/session.go`：

```go
package web

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
)

const sessionCookie = "clash_session"

type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]string // sessionID -> username
}

func NewSessionStore() *SessionStore {
	return &SessionStore{sessions: map[string]string{}}
}

func (s *SessionStore) Create(username string) string {
	id := newID()
	s.mu.Lock()
	s.sessions[id] = username
	s.mu.Unlock()
	return id
}

func (s *SessionStore) Get(id string) (string, bool) {
	s.mu.RLock()
	u, ok := s.sessions[id]
	s.mu.RUnlock()
	return u, ok
}

func (s *SessionStore) Delete(id string) {
	s.mu.Lock()
	delete(s.sessions, id)
	s.mu.Unlock()
}

func newID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
```

创建 `backend-go/internal/web/auth.go`：

```go
package web

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// isAjax 判断是否为前端 AJAX 请求（对齐 Spring Security 的判定逻辑）。
func isAjax(r *http.Request) bool {
	if strings.EqualFold(r.Header.Get("X-Requested-With"), "XMLHttpRequest") {
		return true
	}
	return strings.Contains(r.Header.Get("Accept"), "application/json")
}

// requireAuth 鉴权中间件：未认证时 AJAX 返回 401 JSON，否则 302 跳转 /login。
// 独立函数（不依赖 Handler），由 Task 6 的路由以 requireAuth(h.sessions) 挂载。
func requireAuth(sessions *SessionStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := c.Cookie(sessionCookie)
		if err == nil {
			if _, ok := sessions.Get(id); ok {
				c.Next()
				return
			}
		}
		if isAjax(c.Request) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		c.Redirect(http.StatusFound, "/login")
		c.Abort()
	}
}
```

- [ ] **Step 5: 运行测试确认通过**

Run: `cd backend-go && go test ./internal/service/ -run 'TestEnsureAdmin|TestChangePwd' -v`
Expected: PASS。（`web` 包的中间件由 Task 6 的端到端测试覆盖。）

- [ ] **Step 6: 提交**

```bash
git add backend-go/internal/service backend-go/internal/web
git commit -m "feat(backend-go): add user service, session store and auth middleware

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 6: Web handlers + 路由 + 静态托管

**Files:**
- Create: `backend-go/internal/web/handler.go`
- Create: `backend-go/internal/web/router.go`
- Create: `backend-go/internal/web/static.go`
- Create: `backend-go/internal/web/static/index.html`（占位，构建时同步前端产物）
- Test: `backend-go/internal/web/handler_test.go`

**Interfaces:**
- Consumes: `store.Store`、`service.ClashConfigService`、`service.MergeService`、`service.UserService`、`merge`、`web.SessionStore`（Task 1/3/4/5）、`service.ErrBadOldPassword`/`ErrNotFound`。
- Produces:
  - `web.NewHandler(s *store.Store, cc *service.ClashConfigService, ms *service.MergeService, us *service.UserService) *Handler`
  - `(*Handler).Router() *gin.Engine`

- [ ] **Step 1: 写失败测试**（端到端，`httptest` 驱动 Gin）

创建 `backend-go/internal/web/handler_test.go`：

```go
package web_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/8003901/clash-configs/backend-go/internal/service"
	"github.com/8003901/clash-configs/backend-go/internal/store"
	"github.com/8003901/clash-configs/backend-go/internal/web"
)

const tmpl = `{"proxies":[],"proxy-groups":[],"rules":[]}`

func newRouter(t *testing.T) *httptest.Server {
	t.Helper()
	db, _ := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "t.db")), &gorm.Config{})
	s := store.New(db)
	_ = s.Migrate()
	us := service.NewUserService(s)
	_ = us.EnsureAdmin(false)
	cc := service.NewClashConfigService(s)
	ms := service.NewMergeService(s, tmpl)
	h := web.NewHandler(s, cc, ms, us)
	return httptest.NewServer(h.Router())
}

func TestLoginFlow(t *testing.T) {
	ts := newRouter(t)
	defer ts.Close()
	client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse // 不跟随重定向
	}}

	// 未认证 AJAX 请求 → 401
	req, _ := http.NewRequest("GET", ts.URL+"/clash_configs", nil)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	resp, _ := client.Do(req)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}

	// 表单登录成功 → 302 + 设置 cookie
	form := url.Values{"username": {"admin"}, "password": {"password"}}
	resp, _ = client.PostForm(ts.URL+"/login", form)
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("expected 302, got %d", resp.StatusCode)
	}
	cookies := resp.Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected session cookie")
	}

	// 携带 cookie 访问受保护接口 → 200
	req, _ = http.NewRequest("GET", ts.URL+"/clash_configs", nil)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	for _, ck := range cookies {
		req.AddCookie(ck)
	}
	resp, _ = client.Do(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 after login, got %d", resp.StatusCode)
	}
}

func TestConfigsTokenEndpoint(t *testing.T) {
	ts := newRouter(t)
	defer ts.Close()
	client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}}

	// 未登录也可创建 merge（先直接走 service 层验证 token 端点）
	db, _ := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "t.db")), &gorm.Config{})
	s := store.New(db)
	_ = s.Migrate()
	us := service.NewUserService(s)
	_ = us.EnsureAdmin(false)
	ms := service.NewMergeService(s, tmpl)
	m, _ := ms.Create("m", nil)

	req, _ := http.NewRequest("GET", ts.URL+"/configs?token="+m.Token, nil)
	resp, _ := client.Do(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "text/plain") {
		t.Fatalf("expected text/plain, got %q", ct)
	}

	req, _ = http.NewRequest("GET", ts.URL+"/configs?token=missing", nil)
	resp, _ = client.Do(req)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestChangePasswordEndpoint(t *testing.T) {
	ts := newRouter(t)
	defer ts.Close()
	client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	resp, _ := client.PostForm(ts.URL+"/login", url.Values{"username": {"admin"}, "password": {"password"}})
	cookies := resp.Cookies()

	body := `{"oldPassword":"password","newPassword":"newpass"}`
	req, _ := http.NewRequest("PUT", ts.URL+"/users/me/password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for _, ck := range cookies {
		req.AddCookie(ck)
	}
	resp, _ = client.Do(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	// 旧密码错误 → 400
	body = `{"oldPassword":"wrong","newPassword":"x"}`
	req, _ = http.NewRequest("PUT", ts.URL+"/users/me/password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for _, ck := range cookies {
		req.AddCookie(ck)
	}
	resp, _ = client.Do(req)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestCreateConfigEndpointValidation(t *testing.T) {
	ts := newRouter(t)
	defer ts.Close()
	client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	resp, _ := client.PostForm(ts.URL+"/login", url.Values{"username": {"admin"}, "password": {"password"}})
	cookies := resp.Cookies()

	// 缺 name → 400
	body := `{"url":"https://x","updateSchedule":"DAY","enabled":true}`
	req, _ := http.NewRequest("POST", ts.URL+"/clash_configs", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for _, ck := range cookies {
		req.AddCookie(ck)
	}
	resp, _ = client.Do(req)
	if resp.StatusCode != http.StatusBadRequest {
		var out any
		_ = json.NewDecoder(resp.Body).Decode(&out)
		t.Fatalf("expected 400, got %d (%v)", resp.StatusCode, out)
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd backend-go && go test ./internal/web/ -run TestLoginFlow -v`
Expected: FAIL（`web.NewHandler` 未定义）。

- [ ] **Step 3: 实现静态嵌入**

创建 `backend-go/internal/web/static.go`：

```go
package web

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed static
var staticFS embed.FS

// staticSub 返回以 static/ 为根的文件系统（含 index.html、assets/、favicon.svg、icons.svg）。
func staticSub() http.FileSystem {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(err)
	}
	return http.FS(sub)
}

func registerStatic(r *gin.Engine) {
	// http.FileServer 把 URL 路径映射到 static/ 下的文件：
	//   /              -> static/index.html（目录默认返回 index.html）
	//   /assets/x.js   -> static/assets/x.js
	//   /favicon.svg   -> static/favicon.svg
	fileServer := gin.WrapH(http.FileServer(staticSub()))
	r.GET("/", fileServer)
	r.GET("/index.html", fileServer)
	r.GET("/assets/*filepath", fileServer)
	r.GET("/favicon.svg", fileServer)
	r.GET("/icons.svg", fileServer)
}
```

- [ ] **Step 4: 实现 handler**

创建 `backend-go/internal/web/handler.go`：

```go
package web

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/8003901/clash-configs/backend-go/internal/merge"
	"github.com/8003901/clash-configs/backend-go/internal/service"
	"github.com/8003901/clash-configs/backend-go/internal/store"
	"github.com/8003901/clash-configs/backend-go/internal/model"
)

type Handler struct {
	store   *store.Store
	cc      *service.ClashConfigService
	ms      *service.MergeService
	us      *service.UserService
	sessions *SessionStore
}

func NewHandler(s *store.Store, cc *service.ClashConfigService, ms *service.MergeService, us *service.UserService) *Handler {
	return &Handler{store: s, cc: cc, ms: ms, us: us, sessions: NewSessionStore()}
}

// --- auth ---

func (h *Handler) login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	if !h.us.Authenticate(username, password) {
		c.Redirect(http.StatusFound, "/login?error")
		return
	}
	id := h.sessions.Create(username)
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(sessionCookie, id, 0, "/", "", false, true)
	c.Redirect(http.StatusFound, "/")
}

func (h *Handler) logout(c *gin.Context) {
	if id, err := c.Cookie(sessionCookie); err == nil {
		h.sessions.Delete(id)
	}
	c.SetCookie(sessionCookie, "", -1, "/", "", false, true)
	c.Status(http.StatusOK)
}

func (h *Handler) changePassword(c *gin.Context) {
	var req struct {
		OldPassword string `json:"oldPassword"`
		NewPassword string `json:"newPassword"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	username, ok := h.currentUsername(c)
	if !ok {
		c.Status(http.StatusUnauthorized)
		return
	}
	if err := h.us.ChangePwd(username, req.NewPassword, req.OldPassword); err != nil {
		if errors.Is(err, service.ErrBadOldPassword) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "旧密码错误"})
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusOK)
}

func (h *Handler) currentUsername(c *gin.Context) (string, bool) {
	id, err := c.Cookie(sessionCookie)
	if err != nil {
		return "", false
	}
	return h.sessions.Get(id)
}

// --- clash_configs ---

type clashConfigAdd struct {
	URL            string `json:"url"`
	Name           string `json:"name"`
	UpdateSchedule string `json:"updateSchedule"`
	Enabled        bool   `json:"enabled"`
}

func (h *Handler) createClashConfig(c *gin.Context) {
	var req clashConfigAdd
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	if req.URL == "" || req.Name == "" {
		c.Status(http.StatusBadRequest)
		return
	}
	if req.UpdateSchedule == "" {
		req.UpdateSchedule = "DAY"
	}
	if req.UpdateSchedule != "DAY" && req.UpdateSchedule != "WEEK" {
		c.Status(http.StatusBadRequest)
		return
	}
	cc := &model.ClashConfig{URL: req.URL, Name: req.Name, Enabled: req.Enabled, UpdateSchedule: req.UpdateSchedule}
	saved, err := h.cc.Save(cc)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, saved)
}

func (h *Handler) updateClashConfig(c *gin.Context) {
	var req clashConfigAdd
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	if req.URL == "" || req.Name == "" {
		c.Status(http.StatusBadRequest)
		return
	}
	cc := &model.ClashConfig{ID: c.Param("id"), URL: req.URL, Name: req.Name, Enabled: req.Enabled, UpdateSchedule: req.UpdateSchedule}
	saved, err := h.cc.Save(cc)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, saved)
}

func (h *Handler) detailClashConfig(c *gin.Context) {
	renew := c.Query("renew") == "true"
	cc, err := h.cc.Detail(c.Param("id"), renew)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, cc)
}

func (h *Handler) listClashConfigs(c *gin.Context) {
	list, err := h.cc.All()
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *Handler) deleteClashConfig(c *gin.Context) {
	if err := h.cc.Delete(c.Param("id")); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusOK)
}

// --- clash_configs_merge ---

type mergeAdd struct {
	Name      string   `json:"name"`
	ConfigIDs []string `json:"configIds"`
}

func (h *Handler) createMerge(c *gin.Context) {
	var req mergeAdd
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		c.Status(http.StatusBadRequest)
		return
	}
	m, err := h.ms.Create(req.Name, req.ConfigIDs)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, m)
}

func (h *Handler) listMerges(c *gin.Context) {
	list, err := h.ms.List()
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *Handler) detailMerge(c *gin.Context) {
	m, err := h.ms.Detail(c.Param("id"))
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, m)
}

type mergeUpdate struct {
	Name    string                `json:"name"`
	Config  string                `json:"config"`
	Configs []model.ClashConfig   `json:"configs"`
}

func (h *Handler) updateMerge(c *gin.Context) {
	var req mergeUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	if req.Name == "" || req.Config == "" {
		c.Status(http.StatusBadRequest)
		return
	}
	m := &model.ClashConfigsMerge{ID: c.Param("id"), Name: req.Name, Config: req.Config, Configs: req.Configs}
	saved, err := h.ms.Save(m)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, saved)
}

func (h *Handler) refreshMergeToken(c *gin.Context) {
	m, err := h.ms.RefreshToken(c.Param("id"))
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, m)
}

// --- /configs ---

func (h *Handler) queryConfig(c *gin.Context) {
	token := c.Query("token")
	m, err := h.ms.FindByToken(token)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.Status(http.StatusNotFound)
		} else {
			c.Status(http.StatusInternalServerError)
		}
		return
	}
	sources := make([]merge.Source, 0, len(m.Configs))
	for _, cc := range m.Configs {
		sources = append(sources, merge.Source{Content: cc.Content, UserInfo: cc.SubscriptionUserinfo})
	}
	res, err := merge.Merge(m.Config, sources)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	if res.UserInfo != nil {
		c.Header("subscription-userinfo", res.UserInfo.HeaderValue())
	}
	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(res.YAML))
}
```

- [ ] **Step 5: 实现路由**

创建 `backend-go/internal/web/router.go`：

```go
package web

import (
	"github.com/gin-gonic/gin"
)

func (h *Handler) Router() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	// 公开：静态资源 + 订阅端点 + 登录/登出
	registerStatic(r)
	r.GET("/configs", h.queryConfig)
	r.POST("/login", h.login)
	r.POST("/logout", h.logout)

	// 受保护：业务 API
	api := r.Group("")
	api.Use(requireAuth(h.sessions))
	{
		api.GET("/clash_configs", h.listClashConfigs)
		api.POST("/clash_configs", h.createClashConfig)
		api.GET("/clash_configs/:id", h.detailClashConfig)
		api.PUT("/clash_configs/:id", h.updateClashConfig)
		api.DELETE("/clash_configs/:id", h.deleteClashConfig)

		api.GET("/clash_configs_merge", h.listMerges)
		api.POST("/clash_configs_merge", h.createMerge)
		api.GET("/clash_configs_merge/:id", h.detailMerge)
		api.PUT("/clash_configs_merge/:id", h.updateMerge)
		api.PUT("/clash_configs_merge/:id/token", h.refreshMergeToken)

		api.PUT("/users/me/password", h.changePassword)
	}

	return r
}
```

- [ ] **Step 6: 同步前端产物到静态目录**

```bash
mkdir -p backend-go/internal/web/static
cp -r frontend/dist/* backend-go/internal/web/static/
```

- [ ] **Step 7: 运行测试确认通过**

Run: `cd backend-go && go test ./internal/web/ -v`
Expected: PASS（4 个测试）。

- [ ] **Step 8: 提交**

```bash
git add backend-go/internal/web
git commit -m "feat(backend-go): add web handlers, router and static serving

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 7: 定时刷新调度器

**Files:**
- Create: `backend-go/internal/scheduler/scheduler.go`
- Test: `backend-go/internal/scheduler/scheduler_test.go`

**Interfaces:**
- Consumes: `store.Store`（Task 1）、`service.ClashConfigService`（Task 3，用其 `Renew`）。
- Produces:
  - `scheduler.New(s *store.Store, cc *service.ClashConfigService) *Scheduler`
  - `(*Scheduler).Start()`、`(*Scheduler).Stop()`
  - 可注入的 `RenewDaily()` / `RenewWeekly()`（供测试直接调用）

- [ ] **Step 1: 写失败测试**

创建 `backend-go/internal/scheduler/scheduler_test.go`：

```go
package scheduler_test

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/8003901/clash-configs/backend-go/internal/model"
	"github.com/8003901/clash-configs/backend-go/internal/scheduler"
	"github.com/8003901/clash-configs/backend-go/internal/service"
	"github.com/8003901/clash-configs/backend-go/internal/store"
)

func newScheduler(t *testing.T) (*scheduler.Scheduler, *store.Store) {
	t.Helper()
	db, _ := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "t.db")), &gorm.Config{})
	s := store.New(db)
	_ = s.Migrate()
	cc := service.NewClashConfigService(s)
	return scheduler.New(s, cc), s
}

func TestRenewDailyUpdatesContent(t *testing.T) {
	sc, s := newScheduler(t)

	var hits int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.Header().Set("subscription-userinfo", "upload=1; download=1; total=100; expire=1")
		_, _ = w.Write([]byte("proxies:\n  - name: refreshed\n"))
	}))
	defer ts.Close()

	_ = s.CreateClashConfig(&model.ClashConfig{ID: "c1", URL: ts.URL, Name: "x", UpdateSchedule: "DAY"})

	if err := sc.RenewDaily(); err != nil {
		t.Fatal(err)
	}
	got, _ := s.FindClashConfig("c1")
	if got.Content == "" || hits != 1 {
		t.Fatalf("expected refreshed content, hits=%d content=%q", hits, got.Content)
	}

	// WEEK 配置不在 daily 范围内
	_ = s.CreateClashConfig(&model.ClashConfig{ID: "c2", URL: ts.URL, Name: "y", UpdateSchedule: "WEEK"})
	if err := sc.RenewDaily(); err != nil {
		t.Fatal(err)
	}
	if hits != 2 {
		t.Fatalf("daily should only hit DAY config, hits=%d", hits)
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd backend-go && go test ./internal/scheduler/ -v`
Expected: FAIL（`scheduler` 包未定义）。

- [ ] **Step 3: 实现调度器**

创建 `backend-go/internal/scheduler/scheduler.go`：

```go
package scheduler

import (
	"log"

	"github.com/robfig/cron/v3"

	"github.com/8003901/clash-configs/backend-go/internal/service"
	"github.com/8003901/clash-configs/backend-go/internal/store"
)

type Scheduler struct {
	store *store.Store
	cc    *service.ClashConfigService
	cron  *cron.Cron
}

func New(s *store.Store, cc *service.ClashConfigService) *Scheduler {
	return &Scheduler{store: s, cc: cc, cron: cron.New(cron.WithSeconds())}
}

// Start 注册每日 2:00 与每周一 1:00 两个定时任务并启动。
func (s *Scheduler) Start() {
	_, _ = s.cron.AddFunc("0 0 2 * * *", func() {
		if err := s.RenewDaily(); err != nil {
			log.Printf("renew daily failed: %v", err)
		}
	})
	_, _ = s.cron.AddFunc("0 0 1 * * MON", func() {
		if err := s.RenewWeekly(); err != nil {
			log.Printf("renew weekly failed: %v", err)
		}
	})
	s.cron.Start()
}

func (s *Scheduler) Stop() { s.cron.Stop() }

// RenewDaily 逐条刷新 DAY 配置，单条失败仅记录并继续（对齐原 daily 的 try/catch）。
func (s *Scheduler) RenewDaily() error {
	configs, err := s.store.ListClashConfigsBySchedule("DAY")
	if err != nil {
		return err
	}
	for i := range configs {
		if err := s.cc.Renew(&configs[i]); err != nil {
			log.Printf("renew config %s failed: %v", configs[i].Name, err)
		}
	}
	return nil
}

// RenewWeekly 刷新 WEEK 配置，遇到错误立即中止（对齐原 weekly 无 try/catch）。
func (s *Scheduler) RenewWeekly() error {
	configs, err := s.store.ListClashConfigsBySchedule("WEEK")
	if err != nil {
		return err
	}
	for i := range configs {
		if err := s.cc.Renew(&configs[i]); err != nil {
			return err
		}
	}
	return nil
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `cd backend-go && go test ./internal/scheduler/ -v`
Expected: PASS。

- [ ] **Step 5: 提交**

```bash
git add backend-go/internal/scheduler
git commit -m "feat(backend-go): add scheduled config refresh

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 8: 配置加载 + 程序入口

**Files:**
- Create: `backend-go/internal/config/config.go`
- Create: `backend-go/internal/config/config-template.json`（从原项目复制）
- Create: `backend-go/cmd/server/main.go`
- Test: `backend-go/internal/config/config_test.go`

**Interfaces:**
- Consumes: `store`、`service`、`web`、`scheduler`（Task 1/3/4/5/6/7）。
- Produces:
  - `config.Load() (Config, error)`，`Config{ Port string; DBPath string; PwdInit bool }`
  - `config.Template() string`

- [ ] **Step 1: 复制模板文件**

```bash
cp backend/src/main/resources/config-template.json backend-go/internal/config/config-template.json
```

- [ ] **Step 2: 写失败测试**

创建 `backend-go/internal/config/config_test.go`：

```go
package config_test

import (
	"os"
	"testing"

	"github.com/8003901/clash-configs/backend-go/internal/config"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("SERVER_PORT", "")
	t.Setenv("DB_PATH", "")
	t.Setenv("pwdInit", "")
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != "8080" || cfg.DBPath != "data/demo.db" || cfg.PwdInit {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
}

func TestLoadOverrides(t *testing.T) {
	t.Setenv("SERVER_PORT", "8780")
	t.Setenv("DB_PATH", "/tmp/x.db")
	t.Setenv("pwdInit", "true")
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != "8780" || cfg.DBPath != "/tmp/x.db" || !cfg.PwdInit {
		t.Fatalf("unexpected: %+v", cfg)
	}
}

func TestTemplateNonEmpty(t *testing.T) {
	if config.Template() == "" {
		t.Fatal("template should be non-empty")
	}
}
```

- [ ] **Step 3: 实现配置**

创建 `backend-go/internal/config/config.go`：

```go
package config

import (
	_ "embed"
	"os"
)

//go:embed config-template.json
var templateJSON string

type Config struct {
	Port    string
	DBPath  string
	PwdInit bool
}

func Load() (Config, error) {
	cfg := Config{
		Port:    getenv("SERVER_PORT", "8080"),
		DBPath:  getenv("DB_PATH", "data/demo.db"),
		PwdInit: getenv("pwdInit", "false") == "true",
	}
	return cfg, nil
}

func Template() string { return templateJSON }

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `cd backend-go && go test ./internal/config/ -v`
Expected: PASS。

- [ ] **Step 5: 写入口**

创建 `backend-go/cmd/server/main.go`：

```go
package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/8003901/clash-configs/backend-go/internal/config"
	"github.com/8003901/clash-configs/backend-go/internal/scheduler"
	"github.com/8003901/clash-configs/backend-go/internal/service"
	"github.com/8003901/clash-configs/backend-go/internal/store"
	"github.com/8003901/clash-configs/backend-go/internal/web"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if dir := filepath.Dir(cfg.DBPath); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			log.Fatalf("create db dir: %v", err)
		}
	}

	db, err := gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	s := store.New(db)
	if err := s.Migrate(); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	us := service.NewUserService(s)
	if err := us.EnsureAdmin(cfg.PwdInit); err != nil {
		log.Fatalf("ensure admin: %v", err)
	}
	cc := service.NewClashConfigService(s)
	ms := service.NewMergeService(s, config.Template())

	h := web.NewHandler(s, cc, ms, us)

	sc := scheduler.New(s, cc)
	sc.Start()
	defer sc.Stop()

	log.Printf("listening on :%s", cfg.Port)
	if err := h.Router().Run(":" + cfg.Port); err != nil {
		log.Fatalf("run: %v", err)
	}
}
```

- [ ] **Step 6: 全量构建验证**

Run: `cd backend-go && CGO_ENABLED=0 go build ./... && go test ./...`
Expected: 构建成功，全部测试 PASS。

- [ ] **Step 7: 提交**

```bash
git add backend-go/internal/config backend-go/cmd/server
git commit -m "feat(backend-go): add config loading and server entrypoint

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 9: H2 → SQLite 数据迁移

**Files:**
- Create: `backend-go/cmd/migrate/main.go`
- Create: `backend-go/scripts/h2-dump.sh`

**Interfaces:**
- Consumes: `model`、`store`（Task 1）。
- Produces: 可执行迁移命令，把 H2 dump 的 CSV 导入 SQLite。

- [ ] **Step 1: 写 H2 导出脚本**

创建 `backend-go/scripts/h2-dump.sh`（一次性，需 JDK 与 h2 jar）：

```bash
#!/usr/bin/env bash
# 用法: ./scripts/h2-dump.sh <h2-jar路径> <demo.mv.db所在目录> <输出目录>
# 例: ./scripts/h2-dump.sh ~/.gradle/.../h2-2.x.jar ../backend/data ./dump
set -euo pipefail

H2_JAR="$1"
DB_DIR="$2"      # 含 demo.mv.db 的目录（不含 demo.mv.db 后缀）
OUT_DIR="$3"
mkdir -p "$OUT_DIR"

# H2 连接串与凭据来自原 application.yaml
URL="jdbc:h2:file:${DB_DIR}/demo"
USER="clash-configs"
PASS="password1."

for table in clash_configs clash_configs_merge clash_configs_merge_config "user"; do
  echo ">> dumping ${table}"
  java -cp "$H2_JAR" org.h2.tools.Shell \
    -url "$URL" -user "$USER" -password "$PASS" \
    -sql "CALL CSVWRITE('${OUT_DIR}/${table}.csv', 'SELECT * FROM ${table}')"
done
echo ">> done"
```

> 说明：`CSVWRITE` 与 h2 jar 的准确调用方式依赖具体 h2 版本；若 `CALL CSVWRITE` 不可用，改用 `org.h2.tools.Csv` 的 `-write` 选项，或 `SELECT ... FOR UPDATE` 配合 `org.h2.tools.Shell -csv`。执行迁移时按实际 h2 jar 版本微调，保证产出 4 个 CSV 即可。

- [ ] **Step 2: 写迁移命令**

创建 `backend-go/cmd/migrate/main.go`：

```go
package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/8003901/clash-configs/backend-go/internal/model"
)

// 列序与 h2-dump.sh 的 SELECT * 一致（即实体字段定义顺序）。
// 实现时先用 `PRAGMA table_info(...)` 核对 SQLite 实际列名/顺序，再据此微调。
func main() {
	dumpDir := flag.String("dump", "dump", "H2 dump 输出目录（含 4 个 CSV）")
	dbPath := flag.String("out", "data/demo.db", "SQLite 目标路径")
	flag.Parse()

	if dir := filepath.Dir(*dbPath); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			log.Fatalf("mkdir: %v", err)
		}
	}
	db, err := gorm.Open(sqlite.Open(*dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&model.ClashConfig{}, &model.ClashConfigsMerge{}, &model.User{}); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	steps := []struct {
		name string
		fn   func(*gorm.DB, string) error
	}{
		{"clash_configs", importClashConfigs},
		{"user", importUsers},
		{"clash_configs_merge", importMerges},
		{"clash_configs_merge_config", importJoin},
	}
	for _, st := range steps {
		if err := st.fn(db, filepath.Join(*dumpDir, st.name+".csv")); err != nil {
			log.Fatalf("import %s: %v", st.name, err)
		}
	}
	log.Printf("migration complete -> %s", *dbPath)
}

func readCSV(path string) ([][]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return csv.NewReader(f).ReadAll()
}

func parseBool(s string) bool { return s == "TRUE" || s == "true" || s == "1" }

// parseTime 兼容 H2 CSVWRITE 的时间格式（"2006-01-02 15:04:05" 与带毫秒）。
func parseTime(s string) (time.Time, error) {
	for _, layout := range []string{
		"2006-01-02 15:04:05.999",
		"2006-01-02 15:04:05",
		time.RFC3339,
	} {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unparsable time %q", s)
}

func importClashConfigs(db *gorm.DB, path string) error {
	rows, err := readCSV(path)
	if err != nil {
		return err
	}
	for i, r := range rows {
		if i == 0 || len(r) < 9 {
			continue // 跳过表头与短行
		}
		// 列序：ID,URL,NAME,ENABLED,UPDATE_SCHEDULE,CONTENT,CREATED_AT,UPDATED_AT,SUBSCRIPTION_USERINFO
		c := model.ClashConfig{
			ID: r[0], URL: r[1], Name: r[2], Enabled: parseBool(r[3]),
			UpdateSchedule: r[4], Content: r[5], SubscriptionUserinfo: r[8],
		}
		if c.CreatedAt, err = parseTime(r[6]); err != nil {
			return fmt.Errorf("row %d: %w", i, err)
		}
		if c.UpdatedAt, err = parseTime(r[7]); err != nil {
			return fmt.Errorf("row %d: %w", i, err)
		}
		if err := db.Create(&c).Error; err != nil {
			return err
		}
	}
	return nil
}

func importUsers(db *gorm.DB, path string) error {
	rows, err := readCSV(path)
	if err != nil {
		return err
	}
	for i, r := range rows {
		if i == 0 || len(r) < 3 {
			continue
		}
		// 列序：ID,USERNAME,PASSWORD
		if err := db.Create(&model.User{ID: r[0], Username: r[1], Password: r[2]}).Error; err != nil {
			return err
		}
	}
	return nil
}

func importMerges(db *gorm.DB, path string) error {
	rows, err := readCSV(path)
	if err != nil {
		return err
	}
	for i, r := range rows {
		if i == 0 || len(r) < 6 {
			continue
		}
		// 列序：ID,NAME,TOKEN,CONFIG,CREATED_AT,UPDATED_AT
		m := model.ClashConfigsMerge{ID: r[0], Name: r[1], Token: r[2], Config: r[3]}
		if m.CreatedAt, err = parseTime(r[4]); err != nil {
			return fmt.Errorf("row %d: %w", i, err)
		}
		if m.UpdatedAt, err = parseTime(r[5]); err != nil {
			return fmt.Errorf("row %d: %w", i, err)
		}
		if err := db.Omit("Configs").Create(&m).Error; err != nil {
			return err
		}
	}
	return nil
}

func importJoin(db *gorm.DB, path string) error {
	rows, err := readCSV(path)
	if err != nil {
		return err
	}
	type join struct {
		MergeID  string `gorm:"column:clash_configs_merge_id"`
		ConfigID string `gorm:"column:clash_config_id"`
	}
	for i, r := range rows {
		if i == 0 || len(r) < 2 {
			continue
		}
		// 列序：CLASH_CONFIGS_MERGE_ID,CONFIG_ID（以 h2-dump.sh 实际导出为准）
		if err := db.Table("clash_configs_merge_config").Create(&join{MergeID: r[0], ConfigID: r[1]}).Error; err != nil {
			return err
		}
	}
	return nil
}
```

- [ ] **Step 3: 验证迁移**

```bash
# 1) 导出（需 JDK 与 h2 jar）
cd backend-go && ./scripts/h2-dump.sh <h2.jar> ../backend/data ./dump
# 2) 导入
CGO_ENABLED=0 go run ./cmd/migrate -dump ./dump -out data/demo.db
# 3) 用 sqlite3 或一次查询核对各表行数与 H2 一致
```

Expected: 迁移完成，`clash_configs`/`clash_configs_merge`/`clash_configs_merge_config`/`user` 行数与原 H2 一致；`user.password` 为原 bcrypt 哈希（可用默认密码登录）。

- [ ] **Step 4: 提交**

```bash
git add backend-go/cmd/migrate backend-go/scripts/h2-dump.sh
git commit -m "feat(backend-go): add H2 to SQLite migration command

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 10: Docker + Makefile + docker-compose

**Files:**
- Create: `backend-go/Dockerfile`
- Create: `backend-go/Makefile`
- Modify: `docker-compose.yaml`（根目录）

**Interfaces:**
- Consumes: 全部已完成任务。
- Produces: 可一键构建/运行/部署。

- [ ] **Step 1: 写 Makefile**

创建 `backend-go/Makefile`：

```makefile
.PHONY: sync build test run migrate

# 同步前端构建产物到静态目录（供 go:embed）
sync:
	mkdir -p internal/web/static
	cp -r ../frontend/dist/* internal/web/static/

build: sync
	CGO_ENABLED=0 go build -o bin/clash-configs ./cmd/server

test:
	go test ./...

run: sync
	go run ./cmd/server

migrate: sync
	go run ./cmd/migrate
```

- [ ] **Step 2: 写 Dockerfile**

创建 `backend-go/Dockerfile`：

```dockerfile
# 阶段1：构建
FROM golang:1.26-alpine AS builder
WORKDIR /app
# 先复制依赖清单以利用层缓存
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /clash-configs ./cmd/server

# 阶段2：运行（纯静态二进制，distroless 亦可）
FROM alpine:3.20
WORKDIR /
COPY --from=builder /clash-configs /clash-configs
ENV TZ=Asia/Shanghai
ENV SERVER_PORT=80
VOLUME /data
ENTRYPOINT ["/clash-configs"]
```

- [ ] **Step 3: 更新 docker-compose**

修改根目录 `docker-compose.yaml`，把构建上下文指向 `backend-go`：

```yaml
services:
  clash-configs:
    image: clash-configs:0.0.1-SNAPSHOT
    build:
      context: ./backend-go
      dockerfile: Dockerfile
    ports:
      - "8780:8780"
    volumes:
      - ./backend-go/data:/data
    container_name: clash-configs
    restart: always
    environment:
      SERVER_PORT: "8780"
      DB_PATH: "/data/demo.db"
      TZ: Asia/Shanghai
    cpus: "1"
    mem_limit: 512M
```

- [ ] **Step 4: 本地验证**

```bash
cd backend-go && make build
CGO_ENABLED=0 ./bin/clash-configs &   # 启动，默认 :8080
curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080/          # 期望 200（index.html）
curl -s -o /dev/null -w "%{http_code}\n" -H "X-Requested-With: XMLHttpRequest" http://localhost:8080/clash_configs  # 期望 401
```

Expected: 静态页 200，未认证 API 401。

- [ ] **Step 5: 提交**

```bash
git add backend-go/Dockerfile backend-go/Makefile docker-compose.yaml
git commit -m "feat(backend-go): add Docker, Makefile and compose config

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## 验收清单（全部任务完成后逐项核对）

- [ ] `cd backend-go && go test ./...` 全绿。
- [ ] `CGO_ENABLED=0 go build ./cmd/server` 产出静态二进制（`file` 显示 `statically linked`）。
- [ ] 启动后：`GET /` 返回前端 index.html；未认证 AJAX 访问 `/clash_configs` 返回 401；表单登录后访问返回 200。
- [ ] `GET /configs?token=<有效token>` 返回 YAML 且带 `subscription-userinfo` 头；无效 token 返回 404。
- [ ] 迁移后旧数据完整，admin 用原密码登录成功。
- [ ] `GET /clash_configs` 返回全部配置（含停用）。
