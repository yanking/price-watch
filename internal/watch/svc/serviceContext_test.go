package svc

import (
	"strings"
	"testing"

	"github.com/yanking/price-watch/internal/watch/config"
	"github.com/yanking/price-watch/pkg/log"
)

func TestNewServiceContext_InvalidLogConfig(t *testing.T) {
	tests := []struct {
		name   string
		config config.Config
		errMsg string
	}{
		{
			name: "empty log level",
			config: config.Config{
				Log: log.Config{Level: "", Format: "json", Output: "stdout"},
			},
			errMsg: "level cannot be empty",
		},
		{
			name: "invalid format",
			config: config.Config{
				Log: log.Config{Level: "info", Format: "invalid", Output: "stdout"},
			},
			errMsg: "format must be",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewServiceContext(tt.config)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("error = %q, want containing %q", err, tt.errMsg)
			}
		})
	}
}
