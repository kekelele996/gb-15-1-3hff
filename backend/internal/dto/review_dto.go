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
	Status string `form:"status" binding:"omitempty,oneof=invited accepted declined completed"`
	Page   int    `form:"page" binding:"omitempty,min=1"`
	Size   int    `form:"size" binding:"omitempty,min=1,max=100"`
}
