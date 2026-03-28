package auth

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/yanking/price-watch/internal/auth/application/assembler"
	"github.com/yanking/price-watch/internal/auth/application/service"
	authconfig "github.com/yanking/price-watch/internal/auth/config"
	domainservice "github.com/yanking/price-watch/internal/auth/domain/service"
	"github.com/yanking/price-watch/internal/auth/infrastructure/oauth"
	"github.com/yanking/price-watch/internal/auth/infrastructure/persistence/repository"
	"github.com/yanking/price-watch/internal/auth/infrastructure/token"
	httphandler "github.com/yanking/price-watch/internal/auth/interfaces/http/handler"
	"github.com/yanking/price-watch/internal/auth/interfaces/http/middleware"
	"gorm.io/gorm"
)

// InitModule 初始化 auth 模块，组装所有依赖并注册路由
func InitModule(
	engine *gin.Engine,
	db *gorm.DB,
	redisClient redis.Cmdable,
	cfg authconfig.Config,
	logger *slog.Logger,
) {
	// 基础设施层：仓储
	userRepo := repository.NewUserRepositoryImpl(db)
	bindRepo := repository.NewThirdPartyBindRepositoryImpl(db)

	// 基础设施层：Token 服务
	tokenService := token.NewJWTTokenService(cfg.JWT.Secret, redisClient)

	// 基础设施层：OAuth 策略
	var oauthStrategies []domainservice.OAuthStrategy
	if cfg.OAuth.GitHub.ClientID != "" {
		oauthStrategies = append(oauthStrategies, oauth.NewGitHubOAuthStrategy(
			cfg.OAuth.GitHub.ClientID,
			cfg.OAuth.GitHub.ClientSecret,
			cfg.OAuth.GitHub.RedirectURL,
		))
	}
	if cfg.OAuth.Wechat.ClientID != "" {
		oauthStrategies = append(oauthStrategies, oauth.NewWeChatOAuthStrategy(
			cfg.OAuth.Wechat.ClientID,
			cfg.OAuth.Wechat.ClientSecret,
			cfg.OAuth.Wechat.RedirectURL,
		))
	}

	// 应用层：Assembler
	userAssembler := assembler.NewUserAssembler()

	// 应用层：Service
	authService := service.NewAuthService(userRepo, bindRepo, tokenService, userAssembler, oauthStrategies)
	userService := service.NewUserService(userRepo, userAssembler)

	// 接口层：Handler
	authHandler := httphandler.NewAuthHandler(authService)
	userHandler := httphandler.NewUserHandler(userService)
	authMiddleware := middleware.NewAuthMiddleware(tokenService)

	// 注册路由
	registerRoutes(engine, authHandler, userHandler, authMiddleware)

	logger.Info("auth 模块初始化完成")
}

// registerRoutes 注册路由
func registerRoutes(
	engine *gin.Engine,
	authHandler *httphandler.AuthHandler,
	userHandler *httphandler.UserHandler,
	authMiddleware *middleware.AuthMiddleware,
) {
	v1 := engine.Group("/api/v1")
	{
		// 认证相关路由（无需认证）
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/oauth", authHandler.OAuthLogin)
		}

		// 用户相关路由（需要认证）
		user := v1.Group("/user")
		user.Use(authMiddleware.RequireAuth())
		{
			user.GET("/profile", userHandler.GetProfile)
			user.PUT("/profile", userHandler.UpdateProfile)
			user.PUT("/password", userHandler.ChangePassword)
		}
	}
}
