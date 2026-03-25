package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/yanking/price-watch/internal/watch/exchange"
	"github.com/yanking/price-watch/internal/watch/subscription"
	"github.com/yanking/price-watch/pkg/database/influxdb"
)

func RegisterRoutes(
	e *gin.Engine,
	rds redis.Cmdable,
	influx *influxdb.Client,
	subMgr *subscription.Manager,
	adapters []exchange.ExchangeAdapter,
) {
	h := NewHealthHandler(rds, influx)
	e.GET("/health", h.Health)
	e.GET("/ready", h.Ready)

	v1 := e.Group("/api/v1")
	{
		sub := NewSubscriptionHandler(subMgr)
		v1.POST("/subscriptions", sub.Add)
		v1.DELETE("/subscriptions", sub.Remove)
		v1.GET("/subscriptions", sub.List)

		price := NewPriceHandler(rds)
		v1.GET("/prices", price.ListAll)
		v1.GET("/prices/:symbol", price.GetBySymbol)

		kline := NewKlineHandler(influx, adapters)
		v1.GET("/klines/:symbol", kline.Get)
	}
}
