package mysqlx

import (
	"context"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// sqliteClient 用 SQLite 构造测试用 Client
func sqliteClient(t *testing.T) *Client {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	return newClient(db)
}

func TestNew_InvalidConfig(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		errMsg string
	}{
		{
			name:   "DSN为空",
			config: Config{DSN: ""},
			errMsg: "validate config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(tt.config)
			if err == nil {
				t.Error("New() expected error but got nil")
				if client != nil {
					_ = client.Close()
				}
				return
			}
			if !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("New() error = %v, want containing %q", err, tt.errMsg)
			}
		})
	}
}

func TestClient_DB(t *testing.T) {
	c := sqliteClient(t)
	defer func() { _ = c.Close() }()

	if c.DB() == nil {
		t.Error("DB() returned nil")
	}
}

func TestClient_Ping(t *testing.T) {
	c := sqliteClient(t)
	defer func() { _ = c.Close() }()

	if err := c.Ping(context.Background()); err != nil {
		t.Errorf("Ping() unexpected error: %v", err)
	}
}

func TestClient_Close(t *testing.T) {
	c := sqliteClient(t)

	if err := c.Close(); err != nil {
		t.Errorf("Close() unexpected error: %v", err)
	}

	// Close 后 Ping 应失败
	if err := c.Ping(context.Background()); err == nil {
		t.Error("Ping() after Close() expected error but got nil")
	}
}
