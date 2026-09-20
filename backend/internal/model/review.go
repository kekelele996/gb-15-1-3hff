package model

import "time"

// Review 同行评审任务。
type Review struct {
	ID                   uint       `gorm:"primaryKey" json:"id"`
	PaperID              uint       `gorm:"not null;index" json:"paper_id"`
	Paper                Paper      `gorm:"foreignKey:PaperID" json:"paper,omitempty"`
	ReviewerID           uint       `gorm:"not null;index" json:"reviewer_id"`
	Reviewer             User       `gorm:"foreignKey:ReviewerID" json:"reviewer,omitempty"`
	Status               string     `gorm:"size:32;not null;default:invited;index" json:"status"`
	Decision             string     `gorm:"size:32" json:"decision"`
	Comments             string     `gorm:"type:text" json:"comments"`
	ConfidentialComments string     `gorm:"type:text" json:"confidential_comments"`
	DueDate              *time.Time `json:"due_date"`
	CompletedAt          *time.Time `json:"completed_at"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}
