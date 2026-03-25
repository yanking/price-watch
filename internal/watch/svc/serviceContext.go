package svc

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/yanking/price-watch/internal/watch/config"
	"github.com/yanking/price-watch/pkg/database/mysqlx"
	"github.com/yanking/price-watch/pkg/database/redisx"
	"github.com/yanking/price-watch/pkg/log"
)

// ServiceContext 服务上下文
type ServiceContext struct {
	Config config.Config
	Logger *slog.Logger
	MySQL  *mysqlx.Client
	Redis  *redisx.Client
}

// NewServiceContext 创建服务上下文
func NewServiceContext(c config.Config) (*ServiceContext, error) {
	// 验证配置
	if err := c.Log.Validate(); err != nil {
		return nil, fmt.Errorf("validate log config: %w", err)
	}

	// 创建日志器
	logger, err := log.NewBuilder().FromConfig(&c.Log).Build()
	if err != nil {
		return nil, fmt.Errorf("create logger: %w", err)
	}

	// 创建 MySQL 客户端
	mysqlClient, err := mysqlx.New(c.MySQL)
	if err != nil {
		return nil, fmt.Errorf("create mysql client: %w", err)
	}

	// 创建 Redis 客户端
	redisClient, err := redisx.New(c.Redis)
	if err != nil {
		return nil, fmt.Errorf("create redis client: %w", err)
	}

	return &ServiceContext{
		Config: c,
		Logger: logger,
		MySQL:  mysqlClient,
		Redis:  redisClient,
	}, nil
}

// Close 关闭服务上下文，释放所有资源
func (ctx *ServiceContext) Close() error {
	var errs []error
	if ctx.MySQL != nil {
		if err := ctx.MySQL.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close mysql: %w", err))
		}
	}
	if ctx.Redis != nil {
		if err := ctx.Redis.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close redis: %w", err))
		}
	}
	return errors.Join(errs...)
}
