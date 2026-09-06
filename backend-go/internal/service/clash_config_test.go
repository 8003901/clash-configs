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
