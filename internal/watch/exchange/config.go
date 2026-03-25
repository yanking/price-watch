package exchange

import (
	"fmt"
	"time"
)

type ExchangeConfig struct {
	Enabled       bool          `mapstructure:"enabled"`
	WsURL         string        `mapstructure:"wsUrl"`
	RestURL       string        `mapstructure:"restUrl"`
	ReconnectBase time.Duration `mapstructure:"reconnectBase"`
	ReconnectMax  time.Duration `mapstructure:"reconnectMax"`
	PingInterval  time.Duration `mapstructure:"pingInterval"`
}

func (c *ExchangeConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	if c.WsURL == "" {
		return fmt.Errorf("wsUrl is required")
	}
	if c.RestURL == "" {
		return fmt.Errorf("restUrl is required")
	}
	if c.ReconnectBase == 0 {
		c.ReconnectBase = time.Second
	}
	if c.ReconnectMax == 0 {
		c.ReconnectMax = 60 * time.Second
	}
	if c.PingInterval == 0 {
		c.PingInterval = 30 * time.Second
	}
	return nil
}

type HTTPConfig struct {
	Addr string `mapstructure:"addr"`
}

func (c *HTTPConfig) Validate() error {
	if c.Addr == "" {
		c.Addr = ":8080"
	}
	return nil
}
