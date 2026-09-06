package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/8003901/clash-configs/backend-go/internal/config"
	"github.com/8003901/clash-configs/backend-go/internal/scheduler"
	"github.com/8003901/clash-configs/backend-go/internal/service"
	"github.com/8003901/clash-configs/backend-go/internal/store"
	"github.com/8003901/clash-configs/backend-go/internal/web"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if dir := filepath.Dir(cfg.DBPath); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			log.Fatalf("create db dir: %v", err)
		}
	}

	db, err := gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	s := store.New(db)
	if err := s.Migrate(); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	us := service.NewUserService(s)
	if err := us.EnsureAdmin(cfg.PwdInit); err != nil {
		log.Fatalf("ensure admin: %v", err)
	}
	cc := service.NewClashConfigService(s)
	ms := service.NewMergeService(s, config.Template())

	h := web.NewHandler(s, cc, ms, us)

	sc := scheduler.New(s, cc)
	sc.Start()
	defer sc.Stop()

	log.Printf("listening on :%s", cfg.Port)
	if err := h.Router().Run(":" + cfg.Port); err != nil {
		log.Fatalf("run: %v", err)
	}
}
