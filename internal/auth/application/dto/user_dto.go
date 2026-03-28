package dto

import "time"

// UserResponse 用户信息响应
type UserResponse struct {
	ID            uint64    `json:"id"`
	Username      string    `json:"username"`
	Nickname      string    `json:"nickname,omitempty"`
	Avatar        string    `json:"avatar,omitempty"`
	Email         string    `json:"email,omitempty"`
	AreaCode      string    `json:"area_code,omitempty"`
	Phone         string    `json:"phone,omitempty"`
	MaskedPhone   string    `json:"masked_phone,omitempty"`
	EmailVerified bool      `json:"email_verified"`
	PhoneVerified bool      `json:"phone_verified"`
	Status        string    `json:"status"`
	OAuthProvider string    `json:"oauth_provider,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// UpdateProfileRequest 更新资料请求
type UpdateProfileRequest struct {
	Username string `json:"username" binding:"required"`
	Nickname string `json:"nickname" binding:"omitempty"`
	Email    string `json:"email" binding:"omitempty,email"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}
