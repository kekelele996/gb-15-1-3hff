package dto

// TrendQuery 趋势统计查询参数。
type TrendQuery struct {
	Days int `form:"days" binding:"omitempty,min=1,max=365"`
}

// OverviewResponse 统计面板概览响应。
type OverviewResponse struct {
	Total          int64   `json:"total"`
	Submitted      int64   `json:"submitted"`
	InitialReview  int64   `json:"initial_review"`
	ExternalReview int64   `json:"external_review"`
	Revision       int64   `json:"revision"`
	Accepted       int64   `json:"accepted"`
	Rejected       int64   `json:"rejected"`
	AcceptanceRate float64 `json:"acceptance_rate"`
	AvgReviewDays  float64 `json:"avg_review_days"`
}
