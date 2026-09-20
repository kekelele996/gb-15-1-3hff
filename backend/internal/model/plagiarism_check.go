package model

import "time"

// PlagiarismCheck 查重检测记录。
type PlagiarismCheck struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	PaperID    uint       `gorm:"not null;uniqueIndex;index" json:"paper_id"`
	Paper      Paper      `gorm:"foreignKey:PaperID" json:"-"`
	Similarity float64    `gorm:"default:0" json:"similarity"`
	Status     string     `gorm:"size:32;not null;default:pending" json:"status"`
	Report     string     `gorm:"type:text" json:"report"`
	CheckedAt  *time.Time `json:"checked_at"`
	CreatedAt  time.Time  `json:"created_at"`
}
