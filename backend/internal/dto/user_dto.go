package dto

// RegisterRequest 注册请求。
type RegisterRequest struct {
	Username    string `json:"username" binding:"required,min=3,max=64"`
	Password    string `json:"password" binding:"required,min=6,max=64"`
	Email       string `json:"email" binding:"omitempty,email,max=128"`
	RealName    string `json:"real_name" binding:"required,max=64"`
	Institution string `json:"institution" binding:"omitempty,max=255"`
	Role        string `json:"role" binding:"omitempty,oneof=author reviewer editor"`
}

// LoginRequest 登录请求。
type LoginRequest struct {
	Username string `json:"username" binding:"required,max=64"`
	Password string `json:"password" binding:"required,max=64"`
}

// LoginResponse 登录响应。
type LoginResponse struct {
	Token string      `json:"token"`
	User  UserSummary `json:"user"`
}

// UserSummary 用户摘要。
type UserSummary struct {
	ID          uint   `json:"id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	RealName    string `json:"real_name"`
	Institution string `json:"institution"`
	Role        string `json:"role"`
}

// UpdateProfileRequest 更新资料请求。
type UpdateProfileRequest struct {
	Email       string `json:"email" binding:"omitempty,email,max=128"`
	RealName    string `json:"real_name" binding:"omitempty,max=64"`
	Institution string `json:"institution" binding:"omitempty,max=255"`
}
