package scheduler

import (
	"log"

	"github.com/robfig/cron/v3"

	"github.com/8003901/clash-configs/backend-go/internal/service"
	"github.com/8003901/clash-configs/backend-go/internal/store"
)

type Scheduler struct {
	store *store.Store
	cc    *service.ClashConfigService
	cron  *cron.Cron
}

func New(s *store.Store, cc *service.ClashConfigService) *Scheduler {
	return &Scheduler{store: s, cc: cc, cron: cron.New(cron.WithSeconds())}
}

// Start 注册每日 2:00 与每周一 1:00 两个定时任务并启动。
func (s *Scheduler) Start() {
	_, _ = s.cron.AddFunc("0 0 2 * * *", func() {
		if err := s.RenewDaily(); err != nil {
			log.Printf("renew daily failed: %v", err)
		}
	})
	_, _ = s.cron.AddFunc("0 0 1 * * MON", func() {
		if err := s.RenewWeekly(); err != nil {
			log.Printf("renew weekly failed: %v", err)
		}
	})
	s.cron.Start()
}

func (s *Scheduler) Stop() { s.cron.Stop() }

// RenewDaily 逐条刷新 DAY 配置，单条失败仅记录并继续（对齐原 daily 的 try/catch）。
func (s *Scheduler) RenewDaily() error {
	configs, err := s.store.ListClashConfigsBySchedule("DAY")
	if err != nil {
		return err
	}
	for i := range configs {
		if err := s.cc.Renew(&configs[i]); err != nil {
			log.Printf("renew config %s failed: %v", configs[i].Name, err)
		}
	}
	return nil
}

// RenewWeekly 刷新 WEEK 配置，遇到错误立即中止（对齐原 weekly 无 try/catch）。
func (s *Scheduler) RenewWeekly() error {
	configs, err := s.store.ListClashConfigsBySchedule("WEEK")
	if err != nil {
		return err
	}
	for i := range configs {
		if err := s.cc.Renew(&configs[i]); err != nil {
			return err
		}
	}
	return nil
}
