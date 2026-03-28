package assembler

import (
	"github.com/yanking/price-watch/internal/auth/application/dto"
	"github.com/yanking/price-watch/internal/auth/domain/entity"
)

// UserAssembler 用户装配器
type UserAssembler struct{}

// NewUserAssembler 创建用户装配器
func NewUserAssembler() *UserAssembler {
	return &UserAssembler{}
}

// ToResponse 将实体转换为响应 DTO
func (a *UserAssembler) ToResponse(user *entity.User) *dto.UserResponse {
	if user == nil {
		return nil
	}

	response := &dto.UserResponse{
		ID:            user.ID(),
		Username:      user.Username(),
		Nickname:      user.Nickname(),
		Avatar:        user.Avatar(),
		EmailVerified: user.EmailVerified(),
		PhoneVerified: user.PhoneVerified(),
	Status:        string(user.Status()),
		CreatedAt:     user.CreatedAt(),
		UpdatedAt:     user.UpdatedAt(),
	}

	// 邮箱
	if user.Email() != nil {
		response.Email = user.Email().Value()
	}

	// 手机号区号
	if user.AreaCode() != "" {
		response.AreaCode = user.AreaCode()
	}

	// 手机号
	if user.Phone() != "" {
		response.Phone = user.Phone()
		response.MaskedPhone = user.MaskedPhone()
	}

	return response
}

// ToResponseList 将实体列表转换为响应 DTO 列表
func (a *UserAssembler) ToResponseList(users []*entity.User) []*dto.UserResponse {
	if users == nil {
		return nil
	}

	responses := make([]*dto.UserResponse, 0, len(users))
	for _, user := range users {
		responses = append(responses, a.ToResponse(user))
	}

	return responses
}
