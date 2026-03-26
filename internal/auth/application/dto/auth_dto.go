package dto

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
	Email    string `json:"email" binding:"required,email"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Account  string `json:"account" binding:"required"` // 可以是用户名、邮箱或手机号
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token    string       `json:"token"`
	UserInfo *UserResponse `json:"user_info"`
}

// OAuthLoginRequest OAuth 登录请求
type OAuthLoginRequest struct {
	Code     string `json:"code" binding:"required"`
	Provider string `json:"provider" binding:"required,oneof=github wechat"`
	State    string `json:"state" binding:"required"`
}
