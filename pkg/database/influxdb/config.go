package influxdb

import (
	"fmt"
	"time"
)

type Config struct {
	URL           string        `mapstructure:"url"`
	Token         string        `mapstructure:"token"`
	Org           string        `mapstructure:"org"`
	Buckets       BucketsConfig `mapstructure:"buckets"`
	BatchSize     uint          `mapstructure:"batchSize"`
	FlushInterval time.Duration `mapstructure:"flushInterval"`
}

type BucketsConfig struct {
	Tick  string `mapstructure:"tick"`
	Kline string `mapstructure:"kline"`
}

func (c *Config) Validate() error {
	if c.URL == "" {
		return fmt.Errorf("url is required")
	}
	if c.Org == "" {
		return fmt.Errorf("org is required")
	}
	if c.Buckets.Tick == "" {
		return fmt.Errorf("buckets.tick is required")
	}
	if c.Buckets.Kline == "" {
		return fmt.Errorf("buckets.kline is required")
	}
	if c.BatchSize == 0 {
		c.BatchSize = 500
	}
	if c.FlushInterval == 0 {
		c.FlushInterval = 5 * time.Second
	}
	return nil
}
