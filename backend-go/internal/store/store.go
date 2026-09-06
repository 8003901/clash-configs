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
