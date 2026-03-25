// Package log 提供基于 slog 的结构化日志功能
package log

import "fmt"

// Config 日志配置
type Config struct {
	Level       string `mapstructure:"level"`       // debug, info, warn, error
	Format      string `mapstructure:"format"`      // json, text
	Output      string `mapstructure:"output"`      // stdout, stderr
	EnableTrace bool   `mapstructure:"enableTrace"` // 启用 OpenTelemetry trace
	AddSource   bool   `mapstructure:"addSource"`   // 添加源码位置
	TimeFormat  string `mapstructure:"timeFormat"`  // 时间格式（如 2006-01-02 15:04:05）
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.Level == "" {
		return fmt.Errorf("level cannot be empty")
	}
	if c.Format != "json" && c.Format != "text" {
		return fmt.Errorf("format must be 'json' or 'text'")
	}
	if c.Output != "stdout" && c.Output != "stderr" {
		return fmt.Errorf("output must be 'stdout' or 'stderr'")
	}
	return nil
}
