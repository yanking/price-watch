package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/yanking/price-watch/pkg/database/influxdb"
)

type HealthHandler struct {
	redis  redis.Cmdable
	influx *influxdb.Client
}

func NewHealthHandler(rds redis.Cmdable, influx *influxdb.Client) *HealthHandler {
	return &HealthHandler{redis: rds, influx: influx}
}

func (h *HealthHandler) Health(c *gin.Context) {
	OK(c, gin.H{"status": "alive"})
}

func (h *HealthHandler) Ready(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	if err := h.redis.Ping(ctx).Err(); err != nil {
		Error(c, http.StatusServiceUnavailable, "redis not ready")
		return
	}
	if err := h.influx.Ping(ctx); err != nil {
		Error(c, http.StatusServiceUnavailable, "influxdb not ready")
		return
	}
	OK(c, gin.H{"status": "ready"})
}
