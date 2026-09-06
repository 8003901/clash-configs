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
