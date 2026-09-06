package config

import (
	_ "embed"
	"os"
)

//go:embed config-template.json
var templateJSON string

type Config struct {
	Port    string
	DBPath  string
	PwdInit bool
}

func Load() (Config, error) {
	cfg := Config{
		Port:    getenv("SERVER_PORT", "8080"),
		DBPath:  getenv("DB_PATH", "data/demo.db"),
		PwdInit: getenv("pwdInit", "false") == "true",
	}
	return cfg, nil
}

func Template() string { return templateJSON }

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
