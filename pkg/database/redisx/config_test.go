// pkg/database/redisx/config_test.go
package redisx

import (
	"strings"
	"testing"
	"time"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "有效配置 - 单地址",
			config: Config{
				Addrs: []string{"localhost:6379"},
			},
			wantErr: false,
		},
		{
			name: "有效配置 - 多地址（集群）",
			config: Config{
				Addrs: []string{"node1:6379", "node2:6379", "node3:6379"},
			},
			wantErr: false,
		},
		{
			name: "有效配置 - 完整参数",
			config: Config{
				Addrs:        []string{"localhost:6379"},
				Password:     "secret",
				DB:           1,
				PoolSize:     10,
				MinIdleConns: 5,
				DialTimeout:  5 * time.Second,
				ReadTimeout:  3 * time.Second,
				WriteTimeout: 3 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "有效配置 - DB为零值",
			config: Config{
				Addrs: []string{"localhost:6379"},
				DB:    0,
			},
			wantErr: false,
		},
		{
			name: "有效配置 - 无密码",
			config: Config{
				Addrs:    []string{"localhost:6379"},
				Password: "",
			},
			wantErr: false,
		},
		{
			name:    "无效配置 - Addrs为空",
			config:  Config{},
			wantErr: true,
			errMsg:  "addrs is required",
		},
		{
			name: "无效配置 - Addrs为nil",
			config: Config{
				Addrs: nil,
			},
			wantErr: true,
			errMsg:  "addrs is required",
		},
		{
			name: "无效配置 - Addrs含空字符串",
			config: Config{
				Addrs: []string{"localhost:6379", ""},
			},
			wantErr: true,
			errMsg:  "addr at index 1 is empty",
		},
		{
			name: "无效配置 - DB为负数",
			config: Config{
				Addrs: []string{"localhost:6379"},
				DB:    -1,
			},
			wantErr: true,
			errMsg:  "db must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantErr {
				if err == nil {
					t.Error("Validate() expected error but got nil")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error = %v, want containing %q", err, tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("Validate() unexpected error: %v", err)
			}
		})
	}
}
