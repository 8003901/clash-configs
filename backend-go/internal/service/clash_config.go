package service

import (
	"errors"
	"fmt"
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
	if resp.StatusCode >= 400 {
		return fmt.Errorf("subscription fetch returned status %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	cc.Content = string(body)
	cc.SubscriptionUserinfo = resp.Header.Get("subscription-userinfo")
	return nil
}
