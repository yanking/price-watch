package mysqlx

import (
	"fmt"
	"time"
)

// Config MySQL 连接配置
type Config struct {
	DSN          string        `mapstructure:"dsn"`
	MaxIdleConns int           `mapstructure:"maxIdleConns"`
	MaxOpenConns int           `mapstructure:"maxOpenConns"`
	MaxLifetime  time.Duration `mapstructure:"maxLifetime"`
}

// Validate 校验配置合法性
func (c *Config) Validate() error {
	if c.DSN == "" {
		return fmt.Errorf("dsn is required")
	}
	return nil
}
