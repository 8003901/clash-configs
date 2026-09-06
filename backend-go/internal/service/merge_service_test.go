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
