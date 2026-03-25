package mysqlx

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
			name: "有效配置 - 完整参数",
			config: Config{
				DSN:          "user:pass@tcp(localhost:3306)/db",
				MaxIdleConns: 10,
				MaxOpenConns: 100,
				MaxLifetime:  30 * time.Minute,
			},
			wantErr: false,
		},
		{
			name: "有效配置 - 仅DSN",
			config: Config{
				DSN: "user:pass@tcp(localhost:3306)/db",
			},
			wantErr: false,
		},
		{
			name: "无效配置 - DSN为空",
			config: Config{
				DSN: "",
			},
			wantErr: true,
			errMsg:  "dsn is required",
		},
		{
			name: "有效配置 - 负数MaxOpenConns透传",
			config: Config{
				DSN:          "user:pass@tcp(localhost:3306)/db",
				MaxOpenConns: -1,
			},
			wantErr: false,
		},
		{
			name: "有效配置 - 零值连接池参数",
			config: Config{
				DSN:          "user:pass@tcp(localhost:3306)/db",
				MaxIdleConns: 0,
				MaxOpenConns: 0,
				MaxLifetime:  0,
			},
			wantErr: false,
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
