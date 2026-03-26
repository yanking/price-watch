package token

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/yanking/price-watch/internal/auth/domain/entity"
	"github.com/yanking/price-watch/internal/auth/domain/service"
)

const (
	// TokenVersionKeyPrefix Token 版本 Redis 键前缀
	TokenVersionKeyPrefix = "auth:token:version:"
	// TokenExpire Token 默认过期时间（7天）
	TokenExpire = 7 * 24 * time.Hour
)

// JWTClaims JWT 声明
type JWTClaims struct {
	UserID  int64  `json:"user_id"`
	Version int64  `json:"version"`
	jwt.RegisteredClaims
}

// JWTTokenService JWT Token 服务实现
type JWTTokenService struct {
	secret  []byte        // JWT 密钥
	redis   redis.Cmdable // Redis 客户端
	expire  time.Duration // Token 过期时间
}

// NewJWTTokenService 创建 JWT Token 服务
func NewJWTTokenService(secret string, redisClient redis.Cmdable) service.TokenService {
	return &JWTTokenService{
		secret: []byte(secret),
		redis:  redisClient,
		expire: TokenExpire,
	}
}

// GenerateToken 生成 JWT Token
func (s *JWTTokenService) GenerateToken(user *entity.User) (string, int64, error) {
	// 获取用户当前 Token 版本
	version, err := s.GetVersion(int64(user.ID()))
	if err != nil {
		return "", 0, fmt.Errorf("获取 token 版本失败: %w", err)
	}

	// 创建 JWT 声明
	now := time.Now()
	claims := JWTClaims{
		UserID:  int64(user.ID()),
		Version: version,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.expire)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	// 生成 Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secret)
	if err != nil {
		return "", 0, fmt.Errorf("生成 token 失败: %w", err)
	}

	return tokenString, version, nil
}

// ParseToken 解析 Token
func (s *JWTTokenService) ParseToken(tokenString string) (int64, error) {
	// 解析 Token
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("无效的签名算法: %v", token.Header["alg"])
		}
		return s.secret, nil
	})

	if err != nil {
		return 0, fmt.Errorf("解析 token 失败: %w", err)
	}

	// 提取声明
	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return 0, errors.New("无效的 token")
	}

	// 验证 Token 版本
	currentVersion, err := s.GetVersion(claims.UserID)
	if err != nil {
		return 0, fmt.Errorf("获取 token 版本失败: %w", err)
	}

	if claims.Version != currentVersion {
		return 0, errors.New("token 已失效，请重新登录")
	}

	return claims.UserID, nil
}

// IncrementVersion 递增用户 Token 版本
func (s *JWTTokenService) IncrementVersion(userId int64) (int64, error) {
	key := s.getTokenVersionKey(userId)

	// 使用 Redis INCR 原子递增
	result, err := s.redis.Incr(context.Background(), key).Result()
	if err != nil {
		return 0, fmt.Errorf("递增 token 版本失败: %w", err)
	}

	// 设置过期时间（30天）
	if err := s.redis.Expire(context.Background(), key, 30*24*time.Hour).Err(); err != nil {
		return 0, fmt.Errorf("设置 token 版本过期时间失败: %w", err)
	}

	return result, nil
}

// GetVersion 获取用户当前 Token 版本
func (s *JWTTokenService) GetVersion(userId int64) (int64, error) {
	key := s.getTokenVersionKey(userId)

	// 从 Redis 获取版本号
	result, err := s.redis.Get(context.Background(), key).Result()
	if err != nil {
		if err == redis.Nil {
			// 不存在则初始化为 0
			if err := s.redis.Set(context.Background(), key, "0", 30*24*time.Hour).Err(); err != nil {
				return 0, fmt.Errorf("初始化 token 版本失败: %w", err)
			}
			return 0, nil
		}
		return 0, fmt.Errorf("获取 token 版本失败: %w", err)
	}

	var version int64
	if _, err := fmt.Sscanf(result, "%d", &version); err != nil {
		return 0, fmt.Errorf("解析 token 版本失败: %w", err)
	}

	return version, nil
}

// getTokenVersionKey 获取 Token 版本 Redis 键
func (s *JWTTokenService) getTokenVersionKey(userId int64) string {
	return fmt.Sprintf("%s%d", TokenVersionKeyPrefix, userId)
}
