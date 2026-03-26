package http

import (
	"github.com/gin-gonic/gin"
	"github.com/yanking/price-watch/internal/auth/application/service"
	"github.com/yanking/price-watch/internal/auth/interfaces/http/handler"
	"github.com/yanking/price-watch/internal/auth/interfaces/http/middleware"
	domainservice "github.com/yanking/price-watch/internal/auth/domain/service"
)

// Router 路由
type Router struct {
	engine           *gin.Engine
	authMiddleware   *middleware.AuthMiddleware
	authHandler      *handler.AuthHandler
	userHandler      *handler.UserHandler
}

// NewRouter 创建路由
func NewRouter(
	engine *gin.Engine,
	authService *service.AuthService,
	userService *service.UserService,
	tokenService domainservice.TokenService,
) *Router {
	authMiddleware := middleware.NewAuthMiddleware(tokenService)
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)

	return &Router{
		engine:         engine,
		authMiddleware: authMiddleware,
		authHandler:    authHandler,
		userHandler:    userHandler,
	}
}

// Setup 注册路由
func (r *Router) Setup() {
	// API v1 组
	v1 := r.engine.Group("/api/v1")
	{
		// 认证相关路由（无需认证）
		auth := v1.Group("/auth")
		{
			auth.POST("/register", r.authHandler.Register)
			auth.POST("/login", r.authHandler.Login)
			auth.POST("/oauth", r.authHandler.OAuthLogin)
		}

		// 用户相关路由（需要认证）
		user := v1.Group("/user")
		user.Use(r.authMiddleware.RequireAuth())
		{
			user.GET("/profile", r.userHandler.GetProfile)
			user.PUT("/profile", r.userHandler.UpdateProfile)
			user.PUT("/password", r.userHandler.ChangePassword)
		}
	}
}
