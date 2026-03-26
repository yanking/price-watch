package token

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/yanking/price-watch/internal/auth/domain/entity"
	"github.com/yanking/price-watch/internal/auth/domain/valueobject"
)

func TestJWTTokenService_GenerateToken(t *testing.T) {
	// 创建测试 Redis
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("创建 miniredis 失败: %v", err)
	}
	defer mr.Close()

	// 创建 Redis 客户端
	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	// 创建 Token 服务
	tokenService := NewJWTTokenService("test-secret", redisClient)

	// 创建测试用户
	email, err := valueobject.NewEmail("test@example.com")
	if err != nil {
		t.Fatalf("创建邮箱失败: %v", err)
	}
	password, err := valueobject.NewPassword("password123")
	if err != nil {
		t.Fatalf("创建密码失败: %v", err)
	}

	user, err := entity.NewUser("testuser", password, email)
	if err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}
	user.SetID(123)

	// 测试生成 Token
	token, version, err := tokenService.GenerateToken(user)
	if err != nil {
		t.Errorf("生成 token 失败: %v", err)
	}
	if token == "" {
		t.Error("token 不应为空")
	}
	if version != 0 {
		t.Errorf("初始版本应为 0，得到 %d", version)
	}
}

func TestJWTTokenService_ParseToken(t *testing.T) {
	// 创建测试 Redis
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("创建 miniredis 失败: %v", err)
	}
	defer mr.Close()

	// 创建 Redis 客户端
	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	// 创建 Token 服务
	tokenService := NewJWTTokenService("test-secret", redisClient)

	// 创建测试用户
	email, err := valueobject.NewEmail("test@example.com")
	if err != nil {
		t.Fatalf("创建邮箱失败: %v", err)
	}
	password, err := valueobject.NewPassword("password123")
	if err != nil {
		t.Fatalf("创建密码失败: %v", err)
	}

	user, err := entity.NewUser("testuser", password, email)
	if err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}
	user.SetID(123)

	// 生成 Token
	token, _, err := tokenService.GenerateToken(user)
	if err != nil {
		t.Fatalf("生成 token 失败: %v", err)
	}

	// 测试解析 Token
	userId, err := tokenService.ParseToken(token)
	if err != nil {
		t.Errorf("解析 token 失败: %v", err)
	}
	if userId != 123 {
		t.Errorf("期望用户 ID 123，得到 %d", userId)
	}

	// 测试无效 Token
	_, err = tokenService.ParseToken("invalid-token")
	if err == nil {
		t.Error("期望解析无效 token 返回错误")
	}
}

func TestJWTTokenService_IncrementVersion(t *testing.T) {
	// 创建测试 Redis
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("创建 miniredis 失败: %v", err)
	}
	defer mr.Close()

	// 创建 Redis 客户端
	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	// 创建 Token 服务
	tokenService := NewJWTTokenService("test-secret", redisClient)

	// 测试递增版本
	version1, err := tokenService.IncrementVersion(123)
	if err != nil {
		t.Errorf("递增版本失败: %v", err)
	}
	if version1 != 1 {
		t.Errorf("第一次递增后版本应为 1，得到 %d", version1)
	}

	version2, err := tokenService.IncrementVersion(123)
	if err != nil {
		t.Errorf("递增版本失败: %v", err)
	}
	if version2 != 2 {
		t.Errorf("第二次递增后版本应为 2，得到 %d", version2)
	}

	// 验证获取版本
	currentVersion, err := tokenService.GetVersion(123)
	if err != nil {
		t.Errorf("获取版本失败: %v", err)
	}
	if currentVersion != 2 {
		t.Errorf("期望版本 2，得到 %d", currentVersion)
	}
}

func TestJWTTokenService_ParseTokenWithInvalidVersion(t *testing.T) {
	// 创建测试 Redis
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("创建 miniredis 失败: %v", err)
	}
	defer mr.Close()

	// 创建 Redis 客户端
	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	// 创建 Token 服务
	tokenService := NewJWTTokenService("test-secret", redisClient)

	// 创建测试用户
	email, err := valueobject.NewEmail("test@example.com")
	if err != nil {
		t.Fatalf("创建邮箱失败: %v", err)
	}
	password, err := valueobject.NewPassword("password123")
	if err != nil {
		t.Fatalf("创建密码失败: %v", err)
	}

	user, err := entity.NewUser("testuser", password, email)
	if err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}
	user.SetID(123)

	// 生成 Token
	token, _, err := tokenService.GenerateToken(user)
	if err != nil {
		t.Fatalf("生成 token 失败: %v", err)
	}

	// 递增版本（使旧 token 失效）
	_, err = tokenService.IncrementVersion(123)
	if err != nil {
		t.Fatalf("递增版本失败: %v", err)
	}

	// 尝试解析旧 Token（应该失败）
	_, err = tokenService.ParseToken(token)
	if err == nil {
		t.Error("期望解析旧版本 token 返回错误")
	}
}
