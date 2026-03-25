package config

import (
	"github.com/yanking/price-watch/pkg/database/mysqlx"
	"github.com/yanking/price-watch/pkg/database/redisx"
	"github.com/yanking/price-watch/pkg/log"
)

// Config 应用配置
type Config struct {
	Log   log.Config    `mapstructure:"log"`
	MySQL mysqlx.Config `mapstructure:"mysql"`
	Redis redisx.Config `mapstructure:"redis"`
}
