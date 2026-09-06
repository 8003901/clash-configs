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
