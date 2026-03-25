package eventbus

import (
	"fmt"
	"time"
)

type Config struct {
	Driver string       `mapstructure:"driver"`
	Memory MemoryConfig `mapstructure:"memory"`
	NATS   NATSConfig   `mapstructure:"nats"`
}

func (c *Config) Validate() error {
	switch c.Driver {
	case "memory":
		return c.Memory.Validate()
	case "nats":
		return c.NATS.Validate()
	default:
		return fmt.Errorf("unsupported eventbus driver: %q", c.Driver)
	}
}

type MemoryConfig struct {
	BufferSize int `mapstructure:"bufferSize"`
}

func (c *MemoryConfig) Validate() error {
	if c.BufferSize <= 0 {
		c.BufferSize = 1024
	}
	return nil
}

type NATSConfig struct {
	URL     string         `mapstructure:"url"`
	Streams []StreamConfig `mapstructure:"streams"`
}

func (c *NATSConfig) Validate() error {
	if c.URL == "" {
		return fmt.Errorf("nats url is required")
	}
	return nil
}

type StreamConfig struct {
	Name     string   `mapstructure:"name"`
	Subjects []string `mapstructure:"subjects"`
	MaxAge   Duration `mapstructure:"maxAge"`
}

type Duration time.Duration

func (d *Duration) UnmarshalText(text []byte) error {
	dur, err := time.ParseDuration(string(text))
	if err != nil {
		return err
	}
	*d = Duration(dur)
	return nil
}

func (d Duration) Std() time.Duration {
	return time.Duration(d)
}
