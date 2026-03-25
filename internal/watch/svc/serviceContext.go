package svc

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/yanking/price-watch/internal/watch/config"
	"github.com/yanking/price-watch/internal/watch/exchange"
	"github.com/yanking/price-watch/internal/watch/subscription"
	"github.com/yanking/price-watch/pkg/database/influxdb"
	"github.com/yanking/price-watch/pkg/database/redisx"
	"github.com/yanking/price-watch/pkg/eventbus"
	"github.com/yanking/price-watch/pkg/log"
)

type ServiceContext struct {
	Config config.Config
	Logger *slog.Logger
	Redis  *redisx.Client
	Influx *influxdb.Client
	Bus    eventbus.Bus

	// Set by initial.App()
	SubMgr   *subscription.Manager
	Adapters []exchange.ExchangeAdapter
}

func NewServiceContext(c config.Config) (*ServiceContext, error) {
	if err := c.Log.Validate(); err != nil {
		return nil, fmt.Errorf("validate log config: %w", err)
	}

	logger, err := log.NewBuilder().FromConfig(&c.Log).Build()
	if err != nil {
		return nil, fmt.Errorf("create logger: %w", err)
	}

	redisClient, err := redisx.New(c.Redis)
	if err != nil {
		return nil, fmt.Errorf("create redis client: %w", err)
	}

	influxClient, err := influxdb.New(c.InfluxDB)
	if err != nil {
		return nil, fmt.Errorf("create influxdb client: %w", err)
	}

	bus, err := eventbus.New(c.EventBus, logger)
	if err != nil {
		return nil, fmt.Errorf("create eventbus: %w", err)
	}

	return &ServiceContext{
		Config: c,
		Logger: logger,
		Redis:  redisClient,
		Influx: influxClient,
		Bus:    bus,
	}, nil
}

func (ctx *ServiceContext) Close() error {
	var errs []error
	if ctx.Bus != nil {
		if err := ctx.Bus.Drain(); err != nil {
			errs = append(errs, fmt.Errorf("drain eventbus: %w", err))
		}
		if err := ctx.Bus.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close eventbus: %w", err))
		}
	}
	if ctx.Influx != nil {
		if err := ctx.Influx.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close influxdb: %w", err))
		}
	}
	if ctx.Redis != nil {
		if err := ctx.Redis.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close redis: %w", err))
		}
	}
	return errors.Join(errs...)
}
