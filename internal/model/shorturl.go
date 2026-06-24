package model

import (
	"time"
)

type ShortURL struct {
	ID         uint      `gorm:"primary_key" json:"id"`
	ShortCode  string    `gorm:"unique;not null;size:12" json:"short_code"`
	Name       string    `gorm:"size:100" json:"name"`       // 短链名称
	Remark     string    `gorm:"size:500" json:"remark"`     // 备注
	LongURL    string    `gorm:"not null;size:2048" json:"long_url"`
	UserID     uint      `gorm:"default:0" json:"user_id"`
	VisitCount uint      `gorm:"default:0" json:"total_visits"`
	ExpireAt   *time.Time `gorm:"null" json:"expire_at"`
	Status     int       `gorm:"default:1" json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type VisitLog struct {
	ID        uint      `gorm:"primary_key"`
	ShortCode string    `gorm:"not null;size:12"`
	IP        string    `gorm:"size:45"`
	UserAgent string    `gorm:"size:512"`
	Referer   string    `gorm:"size:512"`
	CreatedAt time.Time
}

const (
	StatusActive   = 1
	StatusDisabled = 0
	StatusExpired  = 2
)