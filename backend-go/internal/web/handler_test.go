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

func newRouter(t *testing.T) (*httptest.Server, *store.Store) {
	t.Helper()
	db, _ := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "t.db")), &gorm.Config{})
	s := store.New(db)
	_ = s.Migrate()
	us := service.NewUserService(s)
	_ = us.EnsureAdmin(false)
	cc := service.NewClashConfigService(s)
	ms := service.NewMergeService(s, tmpl)
	h := web.NewHandler(s, cc, ms, us)
	return httptest.NewServer(h.Router()), s
}

func TestLoginFlow(t *testing.T) {
	ts, _ := newRouter(t)
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
	ts, s := newRouter(t)
	defer ts.Close()
	client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}}

	// 未登录也可创建 merge（先直接走 service 层验证 token 端点）
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
	ts, _ := newRouter(t)
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
	ts, _ := newRouter(t)
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

func TestDeleteMergeEndpoint(t *testing.T) {
	ts, s := newRouter(t)
	defer ts.Close()
	client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	resp, err := client.PostForm(ts.URL+"/login", url.Values{"username": {"admin"}, "password": {"password"}})
	if err != nil {
		t.Fatal(err)
	}
	cookies := resp.Cookies()

	m, err := service.NewMergeService(s, tmpl).Create("merge", nil)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodDelete, ts.URL+"/clash_configs_merge/"+m.ID, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	req, err = http.NewRequest(http.MethodDelete, ts.URL+"/clash_configs_merge/"+m.ID, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 after deleting merge, got %d", resp.StatusCode)
	}
}

func TestAdminCanCreateUsersAndOthersCannot(t *testing.T) {
	ts, _ := newRouter(t)
	defer ts.Close()
	client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	adminLogin, err := client.PostForm(ts.URL+"/login", url.Values{"username": {"admin"}, "password": {"password"}})
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, ts.URL+"/users", nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, cookie := range adminLogin.Cookies() {
		req.AddCookie(cookie)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("admin list users: expected 200, got %d", resp.StatusCode)
	}
	var users []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		t.Fatalf("decode users: %v", err)
	}
	resp.Body.Close()
	encodedUsers, err := json.Marshal(users)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encodedUsers), "password") {
		t.Fatalf("user list must not expose password hashes: %s", encodedUsers)
	}

	req, err = http.NewRequest(http.MethodPost, ts.URL+"/users", strings.NewReader(`{"username":"alice","password":"secure-pass"}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	for _, cookie := range adminLogin.Cookies() {
		req.AddCookie(cookie)
	}
	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("admin create user: expected 201, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	userLogin, err := client.PostForm(ts.URL+"/login", url.Values{"username": {"alice"}, "password": {"secure-pass"}})
	if err != nil {
		t.Fatal(err)
	}
	req, err = http.NewRequest(http.MethodGet, ts.URL+"/users", nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, cookie := range userLogin.Cookies() {
		req.AddCookie(cookie)
	}
	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("regular user list users: expected 403, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}
