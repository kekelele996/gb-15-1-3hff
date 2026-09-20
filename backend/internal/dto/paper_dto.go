package dto

// CreatePaperRequest 新建投稿请求。
type CreatePaperRequest struct {
	Title       string `json:"title" binding:"required,min=2,max=512"`
	Abstract    string `json:"abstract" binding:"required,min=10"`
	Keywords    string `json:"keywords" binding:"required,max=512"`
	Subject     string `json:"subject" binding:"required"`
	AuthorsMeta string `json:"authors_meta" binding:"required"`
	FileKey     string `json:"file_key" binding:"required"`
	FileName    string `json:"file_name" binding:"required,max=255"`
}

// UpdatePaperRequest 更新论文元信息请求。
type UpdatePaperRequest struct {
	Title       string `json:"title" binding:"omitempty,min=2,max=512"`
	Abstract    string `json:"abstract" binding:"omitempty,min=10"`
	Keywords    string `json:"keywords" binding:"omitempty,max=512"`
	Subject     string `json:"subject" binding:"omitempty"`
	AuthorsMeta string `json:"authors_meta" binding:"omitempty"`
}

// InitialReviewRequest 编辑部初审请求。
type InitialReviewRequest struct {
	Pass       bool   `json:"pass"`
	Reason     string `json:"reason" binding:"omitempty,max=2000"`
	ReviewerID uint   `json:"reviewer_id" binding:"omitempty"`
}

// FinalDecisionRequest 终审决定请求。
type FinalDecisionRequest struct {
	Decision string `json:"decision" binding:"required,oneof=accepted rejected"`
	Comment  string `json:"comment" binding:"omitempty,max=2000"`
}

// ReviseRequest 修稿重投请求。
type ReviseRequest struct {
	FileKey        string `json:"file_key" binding:"required"`
	FileName       string `json:"file_name" binding:"required,max=255"`
	ResponseLetter string `json:"response_letter" binding:"required,min=10"`
}

// AssignReviewerRequest 追加分配审稿人请求。
type AssignReviewerRequest struct {
	ReviewerID uint `json:"reviewer_id" binding:"required"`
}

// PaperQuery 论文列表查询参数。
type PaperQuery struct {
	Status  string `form:"status" binding:"omitempty,oneof=submitted initial_review external_review revision accepted rejected"`
	Keyword string `form:"keyword"`
	Subject string `form:"subject"`
	Page    int    `form:"page" binding:"omitempty,min=1"`
	Size    int    `form:"size" binding:"omitempty,min=1,max=100"`
}

// PageResult 分页响应。
type PageResult struct {
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Size  int   `json:"size"`
	Items any   `json:"items"`
}
