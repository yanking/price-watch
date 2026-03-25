package config

import (
	"github.com/yanking/price-watch/internal/watch/exchange"
	"github.com/yanking/price-watch/pkg/database/influxdb"
	"github.com/yanking/price-watch/pkg/database/redisx"
	"github.com/yanking/price-watch/pkg/eventbus"
	"github.com/yanking/price-watch/pkg/log"
)

type Config struct {
	Log       log.Config                         `mapstructure:"log"`
	Redis     redisx.Config                      `mapstructure:"redis"`
	InfluxDB  influxdb.Config                    `mapstructure:"influxdb"`
	EventBus  eventbus.Config                    `mapstructure:"eventbus"`
	HTTP      exchange.HTTPConfig                `mapstructure:"http"`
	Exchanges map[string]exchange.ExchangeConfig `mapstructure:"exchanges"`
}
