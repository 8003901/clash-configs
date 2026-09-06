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
