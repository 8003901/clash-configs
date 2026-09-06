package config_test

import (
	"testing"

	"github.com/8003901/clash-configs/backend-go/internal/config"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("SERVER_PORT", "")
	t.Setenv("DB_PATH", "")
	t.Setenv("pwdInit", "")
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != "8080" || cfg.DBPath != "data/demo.db" || cfg.PwdInit {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
}

func TestLoadOverrides(t *testing.T) {
	t.Setenv("SERVER_PORT", "8780")
	t.Setenv("DB_PATH", "/tmp/x.db")
	t.Setenv("pwdInit", "true")
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != "8780" || cfg.DBPath != "/tmp/x.db" || !cfg.PwdInit {
		t.Fatalf("unexpected: %+v", cfg)
	}
}

func TestTemplateNonEmpty(t *testing.T) {
	if config.Template() == "" {
		t.Fatal("template should be non-empty")
	}
}
