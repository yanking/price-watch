package influxdb

import (
	"testing"
	"time"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name:    "empty URL",
			cfg:     Config{},
			wantErr: true,
		},
		{
			name: "empty org",
			cfg: Config{
				URL: "http://localhost:8086",
			},
			wantErr: true,
		},
		{
			name: "empty tick bucket",
			cfg: Config{
				URL: "http://localhost:8086",
				Org: "test-org",
			},
			wantErr: true,
		},
		{
			name: "empty kline bucket",
			cfg: Config{
				URL:     "http://localhost:8086",
				Org:     "test-org",
				Buckets: BucketsConfig{Tick: "prices-tick"},
			},
			wantErr: true,
		},
		{
			name: "valid config",
			cfg: Config{
				URL:           "http://localhost:8086",
				Token:         "test-token",
				Org:           "test-org",
				Buckets:       BucketsConfig{Tick: "prices-tick", Kline: "prices-kline"},
				BatchSize:     500,
				FlushInterval: 5 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "defaults applied",
			cfg: Config{
				URL:     "http://localhost:8086",
				Org:     "test-org",
				Buckets: BucketsConfig{Tick: "prices-tick", Kline: "prices-kline"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_Defaults(t *testing.T) {
	cfg := Config{
		URL:     "http://localhost:8086",
		Org:     "test-org",
		Buckets: BucketsConfig{Tick: "prices-tick", Kline: "prices-kline"},
	}
	_ = cfg.Validate()
	if cfg.BatchSize != 500 {
		t.Errorf("BatchSize = %d, want 500", cfg.BatchSize)
	}
	if cfg.FlushInterval != 5*time.Second {
		t.Errorf("FlushInterval = %v, want 5s", cfg.FlushInterval)
	}
}

func TestNew_ValidConfig(t *testing.T) {
	cfg := Config{
		URL:     "http://localhost:8086",
		Token:   "test-token",
		Org:     "test-org",
		Buckets: BucketsConfig{Tick: "prices-tick", Kline: "prices-kline"},
	}
	client, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer client.Close()

	if client.Config().Org != "test-org" {
		t.Errorf("Config().Org = %q, want %q", client.Config().Org, "test-org")
	}
}

func TestNew_InvalidConfig(t *testing.T) {
	_, err := New(Config{})
	if err == nil {
		t.Fatal("New() expected error for empty config")
	}
}
