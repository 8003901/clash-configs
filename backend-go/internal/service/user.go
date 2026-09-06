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
