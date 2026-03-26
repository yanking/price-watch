package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yanking/price-watch/internal/auth/domain/service"
	"github.com/yanking/price-watch/internal/auth/interfaces/http/response"
)

const (
	// UserIDKey 用户ID在上下文中的key
	UserIDKey = "user_id"
)

// AuthMiddleware 认证中间件
type AuthMiddleware struct {
	tokenService service.TokenService
}

// NewAuthMiddleware 创建认证中间件
func NewAuthMiddleware(tokenService service.TokenService) *AuthMiddleware {
	return &AuthMiddleware{
		tokenService: tokenService,
	}
}

// RequireAuth 需要认证的中间件
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Header 获取 Token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "缺少认证令牌")
			c.Abort()
			return
		}

		// 解析 Bearer Token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "无效的认证令牌格式")
			c.Abort()
			return
		}

		token := parts[1]

		// 验证 Token
		userID, err := m.tokenService.ParseToken(token)
		if err != nil {
			response.Unauthorized(c, "认证令牌无效或已过期")
			c.Abort()
			return
		}

		// 将用户ID存入上下文
		c.Set(UserIDKey, userID)
		c.Next()
	}
}

// GetUserID 从上下文获取用户ID
func GetUserID(c *gin.Context) (int64, bool) {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return 0, false
	}

	id, ok := userID.(int64)
	return id, ok
}

// MustGetUserID 从上下文获取用户ID，如果不存在则panic
func MustGetUserID(c *gin.Context) int64 {
	userID, exists := GetUserID(c)
	if !exists {
		panic("user_id not found in context")
	}
	return userID
}
