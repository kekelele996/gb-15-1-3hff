package model

import "time"

// Revision 修稿记录：作者按审稿意见重新提交。
type Revision struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	PaperID        uint      `gorm:"not null;index" json:"paper_id"`
	Paper          Paper     `gorm:"foreignKey:PaperID" json:"-"`
	Version        int       `gorm:"not null" json:"version"`
	FileName       string    `gorm:"size:255" json:"file_name"`
	FileKey        string    `gorm:"size:512" json:"file_key"`
	ResponseLetter string    `gorm:"type:text" json:"response_letter"`
	SubmittedByID  uint      `gorm:"not null" json:"submitted_by_id"`
	SubmittedBy    User      `gorm:"foreignKey:SubmittedByID" json:"submitted_by,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}
