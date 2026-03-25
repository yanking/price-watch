package svc

import (
	"strings"
	"testing"

	"github.com/yanking/price-watch/internal/watch/config"
	"github.com/yanking/price-watch/pkg/database/mysqlx"
	"github.com/yanking/price-watch/pkg/log"
)

func TestNewServiceContext(t *testing.T) {
	tests := []struct {
		name    string
		config  config.Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "无效配置 - 空日志级别",
			config: config.Config{
				Log: log.Config{
					Level:       "",
					Format:      "json",
					Output:      "stdout",
					EnableTrace: false,
					AddSource:   false,
					TimeFormat:  "2006-01-02 15:04:05",
				},
			},
			wantErr: true,
			errMsg:  "level cannot be empty",
		},
		{
			name: "无效配置 - 错误的格式",
			config: config.Config{
				Log: log.Config{
					Level:       "info",
					Format:      "invalid",
					Output:      "stdout",
					EnableTrace: false,
					AddSource:   false,
					TimeFormat:  "2006-01-02 15:04:05",
				},
			},
			wantErr: true,
			errMsg:  "format must be 'json' or 'text'",
		},
		{
			name: "无效配置 - 错误的输出",
			config: config.Config{
				Log: log.Config{
					Level:       "info",
					Format:      "json",
					Output:      "invalid",
					EnableTrace: false,
					AddSource:   false,
					TimeFormat:  "2006-01-02 15:04:05",
				},
			},
			wantErr: true,
			errMsg:  "output must be 'stdout' or 'stderr'",
		},
		{
			name: "无效配置 - MySQL DSN为空",
			config: config.Config{
				Log: log.Config{
					Level:       "info",
					Format:      "json",
					Output:      "stdout",
					EnableTrace: false,
					AddSource:   false,
					TimeFormat:  "2006-01-02 15:04:05",
				},
				MySQL: mysqlx.Config{
					DSN: "",
				},
			},
			wantErr: true,
			errMsg:  "create mysql client",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, err := NewServiceContext(tt.config)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewServiceContext() expected error but got nil")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("NewServiceContext() error = %v, want containing %q", err, tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("NewServiceContext() unexpected error: %v", err)
				return
			}

			if ctx == nil {
				t.Error("NewServiceContext() returned nil context")
				return
			}

			if ctx.Logger == nil {
				t.Error("NewServiceContext() Logger is nil")
			}

			if ctx.Config.Log.Level != tt.config.Log.Level {
				t.Errorf("Config.Log.Level = %s, want %s", ctx.Config.Log.Level, tt.config.Log.Level)
			}
		})
	}
}
