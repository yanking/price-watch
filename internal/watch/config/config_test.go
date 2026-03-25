package config

import (
	"testing"

	"github.com/yanking/price-watch/pkg/database/influxdb"
	"github.com/yanking/price-watch/pkg/database/redisx"
	"github.com/yanking/price-watch/pkg/eventbus"
	"github.com/yanking/price-watch/pkg/log"
)

func TestConfigStructure(t *testing.T) {
	c := Config{
		Log: log.Config{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		},
		Redis: redisx.Config{
			Addrs: []string{"localhost:6379"},
		},
		InfluxDB: influxdb.Config{
			URL:     "http://localhost:8086",
			Org:     "test",
			Buckets: influxdb.BucketsConfig{Tick: "prices-tick", Kline: "prices-kline"},
		},
		EventBus: eventbus.Config{
			Driver: "memory",
		},
	}

	if c.Log.Level != "info" {
		t.Errorf("Log.Level = %s, want info", c.Log.Level)
	}
	if len(c.Redis.Addrs) == 0 {
		t.Error("Redis.Addrs should not be empty")
	}
	if c.InfluxDB.URL == "" {
		t.Error("InfluxDB.URL should not be empty")
	}
	if c.EventBus.Driver != "memory" {
		t.Errorf("EventBus.Driver = %s, want memory", c.EventBus.Driver)
	}
}
