package consumer

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/yanking/price-watch/internal/watch/event"
	"github.com/yanking/price-watch/pkg/eventbus"
)

const cacheTTL = 5 * time.Minute

type CacheSubscriber struct {
	redis  redis.Cmdable
	bus    eventbus.Bus
	logger *slog.Logger
	sub    eventbus.Subscription
}

func NewCacheSubscriber(rds redis.Cmdable, bus eventbus.Bus, logger *slog.Logger) *CacheSubscriber {
	return &CacheSubscriber{
		redis:  rds,
		bus:    bus,
		logger: logger.With("component", "cache-subscriber"),
	}
}

func (c *CacheSubscriber) Start() error {
	sub, err := event.SubscribeTick(c.bus, "price.tick.>", c.handleTick)
	if err != nil {
		return fmt.Errorf("subscribe tick: %w", err)
	}
	c.sub = sub
	c.logger.Info("cache subscriber started")
	return nil
}

func (c *CacheSubscriber) Stop() error {
	if c.sub != nil {
		return c.sub.Unsubscribe()
	}
	return nil
}

func (c *CacheSubscriber) handleTick(e event.Event[event.TickData]) error {
	ctx := context.Background()

	// Write to Hash: price:{exchange}:{symbol}
	hashKey := fmt.Sprintf("price:%s:%s", e.Source, e.Subject)
	pipe := c.redis.Pipeline()
	pipe.HSet(ctx, hashKey,
		"price", e.Data.Price.String(),
		"volume", e.Data.Volume.String(),
		"timestamp", strconv.FormatInt(e.Timestamp.UnixMilli(), 10),
	)
	pipe.Expire(ctx, hashKey, cacheTTL)

	// Write to Sorted Set: price:all:{symbol}
	zsetKey := fmt.Sprintf("price:all:%s", e.Subject)
	priceFloat, _ := e.Data.Price.Float64()
	pipe.ZAdd(ctx, zsetKey, redis.Z{
		Score:  priceFloat,
		Member: e.Source,
	})
	pipe.Expire(ctx, zsetKey, cacheTTL)

	if _, err := pipe.Exec(ctx); err != nil {
		c.logger.Warn("cache write failed", "error", err)
		// Log and continue — cache is ephemeral
	}
	return nil
}
