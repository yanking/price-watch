// Package conf provides configuration file parsing with support for multiple
// formats (JSON, YAML, etc.) and environment variable binding.
// Uses viper for advanced configuration management beyond standard library
// capabilities (multi-format support, automatic env, env prefix, key replacer).
package conf

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// MustLoad loads configuration from file to obj. Panic if error.
func MustLoad(file string, obj any) {
	if err := Parse(file, obj); err != nil {
		panic(err)
	}
}

// Parse loads configuration from file to obj. Return error if any.
func Parse(file string, obj any) error {
	v := viper.New()
	v.SetConfigFile(file)

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("read config file error: %w", err)
	}

	v.AutomaticEnv()
	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	if err := v.Unmarshal(obj); err != nil {
		return fmt.Errorf("unable to decode config into struct: %w", err)
	}

	return nil
}
