// pkg/database/redisx/redisx.go
package redisx

import (
	"context"
	"fmt"
	"io"

	"github.com/redis/go-redis/v9"
)

// Client Redis 客户端
type Client struct {
	cmd    redis.Cmdable
	closer io.Closer
}

// New 根据配置创建 Redis 客户端，自动推断单机或集群模式
func New(cfg Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	var cmd redis.Cmdable
	var closer io.Closer

	if len(cfg.Addrs) == 1 {
		// 单机模式
		c := redis.NewClient(&redis.Options{
			Addr:         cfg.Addrs[0],
			Password:     cfg.Password,
			DB:           cfg.DB,
			PoolSize:     cfg.PoolSize,
			MinIdleConns: cfg.MinIdleConns,
			DialTimeout:  cfg.DialTimeout,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
		})
		cmd = c
		closer = c
	} else {
		// 集群模式
		c := redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        cfg.Addrs,
			Password:     cfg.Password,
			PoolSize:     cfg.PoolSize,
			MinIdleConns: cfg.MinIdleConns,
			DialTimeout:  cfg.DialTimeout,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
		})
		cmd = c
		closer = c
	}

	return &Client{cmd: cmd, closer: closer}, nil
}

// Client 返回底层 redis.Cmdable
func (c *Client) Client() redis.Cmdable {
	return c.cmd
}

// Ping 检查 Redis 连通性
func (c *Client) Ping(ctx context.Context) error {
	if err := c.cmd.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("ping redis: %w", err)
	}
	return nil
}

// Close 关闭 Redis 连接
func (c *Client) Close() error {
	if err := c.closer.Close(); err != nil {
		return fmt.Errorf("close redis: %w", err)
	}
	return nil
}
