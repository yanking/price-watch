package service

import (
	"context"
	"fmt"

	"github.com/yanking/price-watch/internal/auth/application/assembler"
	"github.com/yanking/price-watch/internal/auth/application/dto"
	"github.com/yanking/price-watch/internal/auth/domain/entity"
	domainerrors "github.com/yanking/price-watch/internal/auth/domain/errors"
	"github.com/yanking/price-watch/internal/auth/domain/repository"
	"github.com/yanking/price-watch/internal/auth/domain/service"
	domainvo "github.com/yanking/price-watch/internal/auth/domain/valueobject"
)

// AuthService 认证服务
type AuthService struct {
	userRepo            repository.UserRepository
	thirdPartyBindRepo  repository.ThirdPartyBindRepository
	tokenService        service.TokenService
	userAssembler      *assembler.UserAssembler
	oauthStrategyMap   map[string]service.OAuthStrategy
}

// NewAuthService 创建认证服务
func NewAuthService(
	userRepo repository.UserRepository,
	thirdPartyBindRepo repository.ThirdPartyBindRepository,
	tokenService service.TokenService,
	userAssembler *assembler.UserAssembler,
	oauthStrategies []service.OAuthStrategy,
) *AuthService {
	strategyMap := make(map[string]service.OAuthStrategy)
	for _, strategy := range oauthStrategies {
		strategyMap[strategy.GetProviderName()] = strategy
	}

	return &AuthService{
		userRepo:           userRepo,
		thirdPartyBindRepo: thirdPartyBindRepo,
		tokenService:       tokenService,
		userAssembler:     userAssembler,
		oauthStrategyMap:  strategyMap,
	}
}

// Register 注册
func (s *AuthService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.LoginResponse, error) {
	// 检查用户名是否已存在
	exists, err := s.userRepo.ExistsByUsername(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("检查用户名失败: %w", err)
	}
	if exists {
		return nil, domainerrors.ErrUsernameExists
	}

	// 检查邮箱是否已存在
	exists, err = s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("检查邮箱失败: %w", err)
	}
	if exists {
		return nil, domainerrors.ErrEmailExists
	}

	// 创建值对象
	email, err := entity.NewEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("邮箱格式错误: %w", err)
	}

	password, err := entity.NewPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("密码格式错误: %w", err)
	}

	// 创建用户
	user, err := entity.NewUser(req.Username, password, email)
	if err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	// 保存用户
	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, fmt.Errorf("保存用户失败: %w", err)
	}

	// 生成 token
	token, _, err := s.tokenService.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("生成 token 失败: %w", err)
	}

	return &dto.LoginResponse{
		Token:    token,
		UserInfo: s.userAssembler.ToResponse(user),
	}, nil
}

// Login 登录
func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	var user *entity.User
	var err error

	// 尝试按用户名查找
	user, err = s.userRepo.FindByUsername(ctx, req.Account)
	if err != nil || user == nil {
		// 尝试按邮箱查找
		user, err = s.userRepo.FindByEmail(ctx, req.Account)
		if err != nil || user == nil {
			// 尝试按手机号查找（假设没有区号，使用默认区号）
			user, err = s.userRepo.FindByPhone(ctx, "86", req.Account)
			if err != nil || user == nil {
				return nil, domainerrors.ErrAccountPassword
			}
		}
	}

	// 验证密码
	if user.Password() == nil {
		return nil, domainerrors.ErrAccountPassword
	}

	if !user.Password().Verify(req.Password) {
		return nil, domainerrors.ErrAccountPassword
	}

	// 检查用户状态
	if !user.IsActive() {
		return nil, domainerrors.ErrUserSuspended
	}

	// 生成 token
	token, _, err := s.tokenService.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("生成 token 失败: %w", err)
	}

	return &dto.LoginResponse{
		Token:    token,
		UserInfo: s.userAssembler.ToResponse(user),
	}, nil
}

// OAuthLogin OAuth 登录
func (s *AuthService) OAuthLogin(ctx context.Context, req *dto.OAuthLoginRequest) (*dto.LoginResponse, error) {
	// 获取对应的 OAuth 策略
	strategy, ok := s.oauthStrategyMap[req.Provider]
	if !ok {
		return nil, fmt.Errorf("不支持的 OAuth 提供商: %s", req.Provider)
	}

	// 获取 OAuth 用户信息
	oauthUserInfo, err := strategy.GetUserInfo(req.Code)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 转换为值对象
	provider, err := domainvo.ParseOAuthProvider(req.Provider)
	if err != nil {
		return nil, fmt.Errorf("无效的 OAuth 提供商: %w", err)
	}

	// 查找是否已有绑定
	bind, err := s.thirdPartyBindRepo.FindByProvider(ctx, provider, oauthUserInfo.ProviderId)
	if err != nil {
		return nil, fmt.Errorf("查询绑定信息失败: %w", err)
	}

	var user *entity.User

	if bind != nil {
		// 已有绑定，直接登录
		user, err = s.userRepo.FindById(ctx, bind.UserId())
		if err != nil {
			return nil, fmt.Errorf("查询用户失败: %w", err)
		}
		if user == nil {
			return nil, domainerrors.ErrUserNotFound
		}
	} else {
		// 没有绑定，需要创建新用户或绑定已有用户
		if oauthUserInfo.Email != "" {
			// 尝试按邮箱查找已有用户
			user, err = s.userRepo.FindByEmail(ctx, oauthUserInfo.Email)
			if err != nil {
				return nil, fmt.Errorf("查询用户失败: %w", err)
			}
		}

		if user == nil {
			// 创建新用户
			var email *entity.Email
			if oauthUserInfo.Email != "" {
				email, err = entity.NewEmail(oauthUserInfo.Email)
				if err != nil {
					return nil, fmt.Errorf("邮箱格式错误: %w", err)
				}
			}

			// 生成随机用户名
			username := fmt.Sprintf("%s_%s", req.Provider, oauthUserInfo.ProviderId)

			user, err = entity.NewUser(username, nil, email)
			if err != nil {
				return nil, fmt.Errorf("创建用户失败: %w", err)
			}

			// 设置 OAuth 信息
			user.SetOAuthProvider(req.Provider, oauthUserInfo.ProviderId)

			// 保存用户
			if err := s.userRepo.Save(ctx, user); err != nil {
				return nil, fmt.Errorf("保存用户失败: %w", err)
			}
		}

		// 创建第三方绑定
		bind = entity.NewThirdPartyBind(
			int64(user.ID()),
			provider,
			oauthUserInfo.ProviderId,
			oauthUserInfo.ProviderName,
		)

		if err := s.thirdPartyBindRepo.Save(ctx, bind); err != nil {
			return nil, fmt.Errorf("保存绑定信息失败: %w", err)
		}
	}

	// 检查用户状态
	if !user.IsActive() {
		return nil, domainerrors.ErrUserSuspended
	}

	// 生成 token
	token, _, err := s.tokenService.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("生成 token 失败: %w", err)
	}

	return &dto.LoginResponse{
		Token:    token,
		UserInfo: s.userAssembler.ToResponse(user),
	}, nil
}
