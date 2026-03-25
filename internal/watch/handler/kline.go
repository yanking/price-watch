package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yanking/price-watch/internal/watch/exchange"
	"github.com/yanking/price-watch/pkg/database/influxdb"
)

type KlineHandler struct {
	influx   *influxdb.Client
	adapters []exchange.ExchangeAdapter
}

func NewKlineHandler(influx *influxdb.Client, adapters []exchange.ExchangeAdapter) *KlineHandler {
	return &KlineHandler{influx: influx, adapters: adapters}
}

func (h *KlineHandler) Get(c *gin.Context) {
	symbol := c.Param("symbol")
	exc := c.Query("exchange")
	interval := c.DefaultQuery("interval", "1m")
	limitStr := c.DefaultQuery("limit", "500")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 || limit > 5000 {
		limit = 500
	}

	startStr := c.Query("start")
	endStr := c.Query("end")
	end := time.Now()
	start := end.Add(-24 * time.Hour)
	if startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			start = t
		}
	}
	if endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			end = t
		}
	}

	bucket := h.influx.Config().Buckets.Kline
	filter := fmt.Sprintf(`|> filter(fn: (r) => r["symbol"] == "%s")`, symbol)
	if exc != "" {
		filter += fmt.Sprintf(` |> filter(fn: (r) => r["exchange"] == "%s")`, exc)
	}
	query := fmt.Sprintf(`from(bucket: "%s")
		|> range(start: %s, stop: %s)
		|> filter(fn: (r) => r["_measurement"] == "spot_kline")
		|> filter(fn: (r) => r["interval"] == "%s")
		%s
		|> limit(n: %d)`,
		bucket, start.Format(time.RFC3339), end.Format(time.RFC3339), interval, filter, limit)

	result, err := h.influx.Query(c.Request.Context(), query)
	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	var klines []gin.H
	for result.Next() {
		record := result.Record()
		klines = append(klines, gin.H{
			"time":   record.Time(),
			"open":   record.ValueByKey("open"),
			"high":   record.ValueByKey("high"),
			"low":    record.ValueByKey("low"),
			"close":  record.ValueByKey("close"),
			"volume": record.ValueByKey("volume"),
		})
	}

	// Fallback to adapter REST if no data in InfluxDB
	if len(klines) == 0 && exc != "" {
		for _, adapter := range h.adapters {
			if adapter.Name() == exc {
				data, fErr := adapter.FetchKlines(c.Request.Context(), symbol, interval, start, end)
				if fErr != nil {
					Error(c, http.StatusInternalServerError, fErr.Error())
					return
				}
				for _, k := range data {
					klines = append(klines, gin.H{
						"interval": k.Interval,
						"open":     k.Open.String(),
						"high":     k.High.String(),
						"low":      k.Low.String(),
						"close":    k.Close.String(),
						"volume":   k.Volume.String(),
					})
				}
				break
			}
		}
	}

	OK(c, gin.H{"symbol": symbol, "interval": interval, "klines": klines})
}
