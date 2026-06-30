package model

import "time"

// User 用户模型
type User struct {
	ID        uint      `gorm:"primary_key;auto_increment" json:"id"`
	Username  string    `gorm:"unique;not null;size:50" json:"username"`
	Password  string    `gorm:"not null;size:255" json:"-"`
	Email     string    `gorm:"size:100" json:"email"`
	Role      int       `gorm:"default:0" json:"role"`
	Status    int       `gorm:"default:1" json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserRole 用户角色常量
const (
	UserRoleNormal = iota // 0 - 普通用户
	UserRoleAdmin         // 1 - 管理员
)

// UserStatus 用户状态常量
const (
	UserStatusDisabled = iota // 0 - 禁用
	UserStatusActive          // 1 - 正常
)