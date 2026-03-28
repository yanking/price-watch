package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/yanking/price-watch/internal/auth/application/assembler"
	"github.com/yanking/price-watch/internal/auth/application/dto"
	"github.com/yanking/price-watch/internal/auth/domain/entity"
	"github.com/yanking/price-watch/internal/auth/domain/repository"
)

// UserService 用户服务
type UserService struct {
	userRepo       repository.UserRepository
	userAssembler *assembler.UserAssembler
}

// NewUserService 创建用户服务
func NewUserService(
	userRepo repository.UserRepository,
	userAssembler *assembler.UserAssembler,
) *UserService {
	return &UserService{
		userRepo:       userRepo,
		userAssembler: userAssembler,
	}
}

// GetProfile 获取用户资料
func (s *UserService) GetProfile(ctx context.Context, userId int64) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindById(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	if user == nil {
		return nil, errors.New("用户不存在")
	}

	return s.userAssembler.ToResponse(user), nil
}

// UpdateProfile 更新用户资料
func (s *UserService) UpdateProfile(ctx context.Context, userId int64, req *dto.UpdateProfileRequest) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindById(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	if user == nil {
		return nil, errors.New("用户不存在")
	}

	// 更新用户名和昵称
	if err := user.UpdateProfile(req.Username, req.Nickname); err != nil {
		return nil, fmt.Errorf("更新用户资料失败: %w", err)
	}

	// 更新邮箱（如果提供）
	if req.Email != "" {
		// 检查邮箱是否被其他用户使用
		exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
		if err != nil {
			return nil, fmt.Errorf("检查邮箱失败: %w", err)
		}
		if exists {
			// 如果邮箱与当前用户邮箱不同，则报错
			if user.Email() == nil || user.Email().Value() != req.Email {
				return nil, errors.New("邮箱已被其他用户使用")
			}
		}

		email, err := entity.NewEmail(req.Email)
		if err != nil {
			return nil, fmt.Errorf("邮箱格式错误: %w", err)
		}
		user.SetEmail(email)
	}

	// 保存更新
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("保存用户失败: %w", err)
	}

	return s.userAssembler.ToResponse(user), nil
}

// ChangePassword 修改密码
func (s *UserService) ChangePassword(ctx context.Context, userId int64, req *dto.ChangePasswordRequest) error {
	user, err := s.userRepo.FindById(ctx, userId)
	if err != nil {
		return fmt.Errorf("查询用户失败: %w", err)
	}
	if user == nil {
		return errors.New("用户不存在")
	}

	// 创建新密码值对象
	newPassword, err := entity.NewPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("密码格式错误: %w", err)
	}

	// 修改密码
	if err := user.ChangePassword(req.OldPassword, newPassword); err != nil {
		return err
	}

	// 保存更新
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("保存用户失败: %w", err)
	}

	return nil
}
