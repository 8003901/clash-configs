package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/8003901/clash-configs/backend-go/internal/model"
)

// 列序以 h2-dump.sh 的 `SELECT *` 实际导出为准（H2 物理列序为字典序，
// 与 Go 结构体字段定义顺序不同）。实测 H2 INFORMATION_SCHEMA.COLUMNS：
//
//	clash_configs:             ID, CONTENT, CREATED_AT, ENABLED, NAME,
//	                           UPDATE_SCHEDULE, UPDATED_AT, URL, SUBSCRIPTION_USERINFO
//	clash_configs_merge:       ID, CONFIG, CREATED_AT, NAME, TOKEN, UPDATED_AT
//	clash_configs_merge_config:CLASH_CONFIGS_MERGE_ID, CONFIG_ID
//	user:                      ID, PASSWORD, USERNAME
func main() {
	dumpDir := flag.String("dump", "dump", "H2 dump 输出目录（含 4 个 CSV）")
	dbPath := flag.String("out", "data/demo.db", "SQLite 目标路径")
	flag.Parse()

	if dir := filepath.Dir(*dbPath); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			log.Fatalf("mkdir: %v", err)
		}
	}
	db, err := gorm.Open(sqlite.Open(*dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&model.ClashConfig{}, &model.ClashConfigsMerge{}, &model.User{}); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	steps := []struct {
		name string
		fn   func(*gorm.DB, string) error
	}{
		{"clash_configs", importClashConfigs},
		{"user", importUsers},
		{"clash_configs_merge", importMerges},
		{"clash_configs_merge_config", importJoin},
	}
	for _, st := range steps {
		if err := st.fn(db, filepath.Join(*dumpDir, st.name+".csv")); err != nil {
			log.Fatalf("import %s: %v", st.name, err)
		}
	}
	log.Printf("migration complete -> %s", *dbPath)
}

func readCSV(path string) ([][]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return csv.NewReader(f).ReadAll()
}

func parseBool(s string) bool { return s == "TRUE" || s == "true" || s == "1" }

// parseTime 兼容 H2 CSVWRITE 的时间格式（"2006-01-02 15:04:05" 与带毫秒）。
func parseTime(s string) (time.Time, error) {
	for _, layout := range []string{
		"2006-01-02 15:04:05.999",
		"2006-01-02 15:04:05",
		time.RFC3339,
	} {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unparsable time %q", s)
}

func importClashConfigs(db *gorm.DB, path string) error {
	rows, err := readCSV(path)
	if err != nil {
		return err
	}
	for i, r := range rows {
		if i == 0 || len(r) < 9 {
			continue // 跳过表头与短行
		}
		// 列序：ID,CONTENT,CREATED_AT,ENABLED,NAME,UPDATE_SCHEDULE,UPDATED_AT,URL,SUBSCRIPTION_USERINFO
		c := model.ClashConfig{
			ID: r[0], URL: r[7], Name: r[4], Enabled: parseBool(r[3]),
			UpdateSchedule: r[5], Content: r[1], SubscriptionUserinfo: r[8],
		}
		if c.CreatedAt, err = parseTime(r[2]); err != nil {
			return fmt.Errorf("row %d: %w", i, err)
		}
		if c.UpdatedAt, err = parseTime(r[6]); err != nil {
			return fmt.Errorf("row %d: %w", i, err)
		}
		if err := db.Create(&c).Error; err != nil {
			return err
		}
	}
	return nil
}

func importUsers(db *gorm.DB, path string) error {
	rows, err := readCSV(path)
	if err != nil {
		return err
	}
	for i, r := range rows {
		if i == 0 || len(r) < 3 {
			continue
		}
		// 列序：ID,PASSWORD,USERNAME
		if err := db.Create(&model.User{ID: r[0], Username: r[2], Password: r[1]}).Error; err != nil {
			return err
		}
	}
	return nil
}

func importMerges(db *gorm.DB, path string) error {
	rows, err := readCSV(path)
	if err != nil {
		return err
	}
	for i, r := range rows {
		if i == 0 || len(r) < 6 {
			continue
		}
		// 列序：ID,CONFIG,CREATED_AT,NAME,TOKEN,UPDATED_AT
		m := model.ClashConfigsMerge{ID: r[0], Name: r[3], Token: r[4], Config: r[1]}
		if m.CreatedAt, err = parseTime(r[2]); err != nil {
			return fmt.Errorf("row %d: %w", i, err)
		}
		if m.UpdatedAt, err = parseTime(r[5]); err != nil {
			return fmt.Errorf("row %d: %w", i, err)
		}
		if err := db.Omit("Configs").Create(&m).Error; err != nil {
			return err
		}
	}
	return nil
}

func importJoin(db *gorm.DB, path string) error {
	rows, err := readCSV(path)
	if err != nil {
		return err
	}
	type join struct {
		MergeID  string `gorm:"column:clash_configs_merge_id"`
		ConfigID string `gorm:"column:clash_config_id"`
	}
	for i, r := range rows {
		if i == 0 || len(r) < 2 {
			continue
		}
		// 列序：CLASH_CONFIGS_MERGE_ID,CONFIG_ID
		if err := db.Table("clash_configs_merge_config").Create(&join{MergeID: r[0], ConfigID: r[1]}).Error; err != nil {
			return err
		}
	}
	return nil
}
