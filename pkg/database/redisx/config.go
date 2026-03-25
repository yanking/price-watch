// pkg/database/redisx/config.go
package redisx

import (
	"fmt"
	"time"
)

// Config Redis 连接配置
type Config struct {
	Addrs        []string      `mapstructure:"addrs"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db"`
	PoolSize     int           `mapstructure:"poolSize"`
	MinIdleConns int           `mapstructure:"minIdleConns"`
	DialTimeout  time.Duration `mapstructure:"dialTimeout"`
	ReadTimeout  time.Duration `mapstructure:"readTimeout"`
	WriteTimeout time.Duration `mapstructure:"writeTimeout"`
}

// Validate 校验配置合法性
func (c *Config) Validate() error {
	if len(c.Addrs) == 0 {
		return fmt.Errorf("addrs is required")
	}
	for i, addr := range c.Addrs {
		if addr == "" {
			return fmt.Errorf("addr at index %d is empty", i)
		}
	}
	if c.DB < 0 {
		return fmt.Errorf("db must be non-negative, got %d", c.DB)
	}
	return nil
}
