// pkg/database/redisx/redisx_test.go
package redisx

import (
	"context"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
)

func TestNew_InvalidConfig(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		errMsg string
	}{
		{
			name:   "Addrs为空",
			config: Config{},
			errMsg: "validate config",
		},
		{
			name: "DB为负数",
			config: Config{
				Addrs: []string{"localhost:6379"},
				DB:    -1,
			},
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

func TestClient_Standalone(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	defer mr.Close()

	c, err := New(Config{
		Addrs: []string{mr.Addr()},
	})
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}
	defer func() { _ = c.Close() }()

	// Client() 返回非 nil
	if c.Client() == nil {
		t.Error("Client() returned nil")
	}

	// Ping 成功
	if err := c.Ping(context.Background()); err != nil {
		t.Errorf("Ping() unexpected error: %v", err)
	}
}

func TestClient_Close(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	defer mr.Close()

	c, err := New(Config{
		Addrs: []string{mr.Addr()},
	})
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}

	if err := c.Close(); err != nil {
		t.Errorf("Close() unexpected error: %v", err)
	}

	// Close 后 Ping 应失败
	if err := c.Ping(context.Background()); err == nil {
		t.Error("Ping() after Close() expected error but got nil")
	}
}
