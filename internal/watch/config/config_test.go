package config

import (
	"testing"
	"time"

	"github.com/yanking/price-watch/pkg/database/mysqlx"
	"github.com/yanking/price-watch/pkg/database/redisx"
	"github.com/yanking/price-watch/pkg/log"
)

func TestConfigStructure(t *testing.T) {
	c := Config{
		Log: log.Config{
			Level:       "info",
			Format:      "json",
			Output:      "stdout",
			EnableTrace: false,
			AddSource:   false,
			TimeFormat:  "2006-01-02 15:04:05",
		},
		MySQL: mysqlx.Config{
			DSN:          "test:test@tcp(localhost:3306)/test",
			MaxIdleConns: 10,
			MaxOpenConns: 100,
			MaxLifetime:  30 * time.Minute,
		},
		Redis: redisx.Config{
			Addrs:    []string{"localhost:6379"},
			Password: "",
			DB:       0,
		},
	}

	if c.Log.Level != "info" {
		t.Errorf("Log.Level = %s, want info", c.Log.Level)
	}
	if c.MySQL.DSN == "" {
		t.Error("MySQL.DSN should not be empty")
	}
	if len(c.Redis.Addrs) == 0 {
		t.Error("Redis.Addrs should not be empty")
	}
}
