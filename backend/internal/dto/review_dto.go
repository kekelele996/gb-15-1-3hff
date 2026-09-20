package dto

// RespondRequest 审稿人接受/拒绝邀请请求。
type RespondRequest struct {
	Accept bool `json:"accept"`
}

// SubmitReviewRequest 提交评审意见请求。
type SubmitReviewRequest struct {
	Decision             string `json:"decision" binding:"required,oneof=accept minor_revision major_revision reject"`
	Comments             string `json:"comments" binding:"required,min=10,max=5000"`
	ConfidentialComments string `json:"confidential_comments" binding:"omitempty,max=5000"`
}

// ReviewQuery 审稿列表查询参数。
type ReviewQuery struct {
	Status string `form:"status" binding:"omitempty,oneof=invited accepted declined completed expired"`
	Page   int    `form:"page" binding:"omitempty,min=1"`
	Size   int    `form:"size" binding:"omitempty,min=1,max=100"`
}

// ReviewSummaryResponse 论文当前轮次审稿进度汇总（终审门禁依据）。
type ReviewSummaryResponse struct {
	PaperID     uint   `json:"paper_id"`
	Round       int    `json:"round"`
	Completed   int    `json:"completed"`   // 已完成（当前轮有效）
	Pending     int    `json:"pending"`     // 待接受（邀请未回应且未超期）
	Declined    int    `json:"declined"`    // 已拒绝
	Expired     int    `json:"expired"`     // 已超期失效
	InProgress  int    `json:"in_progress"` // 审稿中（已接受未提交）
	CanFinalize bool   `json:"can_finalize"`
	BlockReason string `json:"block_reason"`
}
