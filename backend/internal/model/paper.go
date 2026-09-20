package model

import "time"

// Paper 学术论文投稿。
type Paper struct {
	ID                   uint       `gorm:"primaryKey" json:"id"`
	Title                string     `gorm:"size:512;not null;index" json:"title"`
	Abstract             string     `gorm:"type:text" json:"abstract"`
	Keywords             string     `gorm:"size:512" json:"keywords"`
	Subject              string     `gorm:"size:128;not null;index" json:"subject"`
	AuthorsMeta          string     `gorm:"type:text" json:"authors_meta"`
	FileName             string     `gorm:"size:255" json:"file_name"`
	FileKey              string     `gorm:"size:512" json:"file_key"`
	Status               string     `gorm:"size:32;not null;default:submitted;index" json:"status"`
	Version              int        `gorm:"not null;default:1" json:"version"`
	SubmitterID          uint       `gorm:"not null;index" json:"submitter_id"`
	Submitter            User       `gorm:"foreignKey:SubmitterID" json:"submitter,omitempty"`
	InitialReviewComment string     `gorm:"type:text" json:"initial_review_comment"`
	FinalComment         string     `gorm:"type:text" json:"final_comment"`
	FinalDecision        string     `gorm:"size:32" json:"final_decision"`
	Similarity           float64    `gorm:"default:0" json:"similarity"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
	Reviews              []Review   `gorm:"foreignKey:PaperID" json:"reviews,omitempty"`
	Revisions            []Revision `gorm:"foreignKey:PaperID" json:"revisions,omitempty"`
}

// PaperFilter 论文查询过滤条件。
type PaperFilter struct {
	Status      string
	SubmitterID uint
	Keyword     string
	Subject     string
}

// SubjectCount 学科投稿统计。
type SubjectCount struct {
	Subject string `json:"subject"`
	Count   int64  `json:"count"`
}

// DayCount 每日投稿量统计。
type DayCount struct {
	Day   string `json:"day"`
	Count int64  `json:"count"`
}

// ReviewerLoad 审稿人工作量统计。
type ReviewerLoad struct {
	ReviewerID   uint   `json:"reviewer_id"`
	ReviewerName string `json:"reviewer_name"`
	Total        int64  `json:"total"`
	Completed    int64  `json:"completed"`
}
