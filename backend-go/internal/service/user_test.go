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

func TestCreateUserValidationAndDuplicate(t *testing.T) {
	u := newUserService(t)
	if err := u.EnsureAdmin(false); err != nil {
		t.Fatal(err)
	}

	created, err := u.Create("  alice  ", "secure-pass")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	if created.Username != "alice" || !u.Authenticate("alice", "secure-pass") {
		t.Fatalf("created account did not authenticate: %+v", created)
	}
	if _, err := u.Create("alice", "another-pass"); err != service.ErrUsernameExists {
		t.Fatalf("expected ErrUsernameExists, got %v", err)
	}
	for _, tc := range []struct {
		username string
		password string
	}{
		{"ab", "secure-pass"},
		{"valid-name", "short"},
		{"valid-name", string(make([]byte, 73))},
	} {
		if _, err := u.Create(tc.username, tc.password); err != service.ErrInvalidUser {
			t.Fatalf("expected ErrInvalidUser for username=%q password length=%d, got %v", tc.username, len(tc.password), err)
		}
	}
}
