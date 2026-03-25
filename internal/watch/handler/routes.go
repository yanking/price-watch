package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/yanking/price-watch/pkg/database/influxdb"
)

func RegisterRoutes(e *gin.Engine, rds redis.Cmdable, influx *influxdb.Client) {
	h := NewHealthHandler(rds, influx)
	e.GET("/health", h.Health)
	e.GET("/ready", h.Ready)

	// API v1 routes will be added in Tasks 15b and 15c
	_ = e.Group("/api/v1")
}
