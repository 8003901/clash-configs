package model

import "time"

type ClashConfig struct {
	ID                   string    `gorm:"primaryKey;size:36" json:"id"`
	URL                  string    `json:"url"`
	Name                 string    `json:"name"`
	Enabled              bool      `json:"enabled"`
	UpdateSchedule       string    `json:"updateSchedule"`
	Content              string    `json:"content"`
	CreatedAt            time.Time `json:"createdAt"`
	UpdatedAt            time.Time `json:"updatedAt"`
	SubscriptionUserinfo string    `json:"subscriptionUserinfo"`
}

type ClashConfigsMerge struct {
	ID        string        `gorm:"primaryKey;size:36" json:"id"`
	Name      string        `json:"name"`
	Token     string        `json:"token"`
	Config    string        `json:"config"`
	CreatedAt time.Time     `json:"createdAt"`
	UpdatedAt time.Time     `json:"updatedAt"`
	Configs   []ClashConfig `gorm:"many2many:clash_configs_merge_config" json:"configs"`
}

type User struct {
	ID       string `gorm:"primaryKey;size:36" json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}
