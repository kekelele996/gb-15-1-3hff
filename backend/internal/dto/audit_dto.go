package dto

// AuditLogResponse 审计日志响应。
type AuditLogResponse struct {
	ID        uint   `json:"id"`
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	Action    string `json:"action"`
	Entity    string `json:"entity"`
	EntityID  string `json:"entity_id"`
	Detail    string `json:"detail"`
	IP        string `json:"ip"`
	RequestID string `json:"request_id"`
	CreatedAt string `json:"created_at"`
}

// AuditQuery 审计日志查询参数。
type AuditQuery struct {
	Page int `form:"page" binding:"omitempty,min=1"`
	Size int `form:"size" binding:"omitempty,min=1,max=100"`
}
