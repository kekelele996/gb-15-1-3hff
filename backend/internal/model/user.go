package model

import "time"

// User 用户：作者/审稿人/编辑/管理员。
type User struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Username    string    `gorm:"size:64;uniqueIndex;not null" json:"username"`
	Password    string    `gorm:"size:255;not null" json:"-"`
	Email       string    `gorm:"size:128" json:"email"`
	RealName    string    `gorm:"size:64" json:"real_name"`
	Institution string    `gorm:"size:255" json:"institution"`
	Role        string    `gorm:"size:32;not null;default:author;index" json:"role"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
