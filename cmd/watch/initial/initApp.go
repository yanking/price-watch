package initial

import (
	"context"

	"github.com/yanking/price-watch/internal/watch/consumer"
	"github.com/yanking/price-watch/internal/watch/exchange"
	"github.com/yanking/price-watch/internal/watch/subscription"
	"github.com/yanking/price-watch/internal/watch/svc"
)

func App(ctx *svc.ServiceContext) {
	// Create exchange adapters from config
	adapterMap := make(map[string]exchange.ExchangeAdapter)
	for name, cfg := range ctx.Config.Exchanges {
		if !cfg.Enabled {
			continue
		}
		adapter := exchange.NewAdapter(name, ctx.Bus, ctx.Logger, cfg)
		if adapter != nil {
			adapterMap[name] = adapter
			ctx.Adapters = append(ctx.Adapters, adapter)
		}
	}

	// Create SubscriptionManager
	ctx.SubMgr = subscription.NewManager(ctx.Redis.Client(), adapterMap, ctx.Logger)

	// Start CacheSubscriber
	cacheSub := consumer.NewCacheSubscriber(ctx.Redis.Client(), ctx.Bus, ctx.Logger)
	if err := cacheSub.Start(); err != nil {
		ctx.Logger.Error("start cache subscriber", "error", err)
	}

	// Start StorageSubscriber
	storageSub := consumer.NewStorageSubscriber(
		ctx.Influx, ctx.Bus, ctx.Logger,
		int(ctx.Config.InfluxDB.BatchSize),
		ctx.Config.InfluxDB.FlushInterval,
	)
	if err := storageSub.Start(); err != nil {
		ctx.Logger.Error("start storage subscriber", "error", err)
	}

	// Restore subscriptions from Redis
	if err := ctx.SubMgr.Restore(context.Background()); err != nil {
		ctx.Logger.Error("restore subscriptions", "error", err)
	}
}
