package model

import "time"

// AuditLog 操作审计日志。
type AuditLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index" json:"user_id"`
	Username  string    `gorm:"size:64;index" json:"username"`
	Action    string    `gorm:"size:64;index" json:"action"`
	Entity    string    `gorm:"size:64" json:"entity"`
	EntityID  string    `gorm:"size:64" json:"entity_id"`
	Detail    string    `gorm:"type:text" json:"detail"`
	IP        string    `gorm:"size:64" json:"ip"`
	RequestID string    `gorm:"size:64" json:"request_id"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
}
