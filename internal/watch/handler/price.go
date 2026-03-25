package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type PriceHandler struct {
	redis redis.Cmdable
}

func NewPriceHandler(rds redis.Cmdable) *PriceHandler {
	return &PriceHandler{redis: rds}
}

func (h *PriceHandler) GetBySymbol(c *gin.Context) {
	symbol := c.Param("symbol")
	key := fmt.Sprintf("price:all:%s", symbol)
	members, err := h.redis.ZRangeWithScores(c.Request.Context(), key, 0, -1).Result()
	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	type ExchangePrice struct {
		Exchange string `json:"exchange"`
		Price    string `json:"price"`
	}
	prices := make([]ExchangePrice, 0, len(members))
	for _, m := range members {
		prices = append(prices, ExchangePrice{
			Exchange: m.Member.(string),
			Price:    fmt.Sprintf("%g", m.Score),
		})
	}
	OK(c, gin.H{"symbol": symbol, "prices": prices})
}

func (h *PriceHandler) ListAll(c *gin.Context) {
	var cursor uint64
	var allPrices []gin.H

	for {
		keys, nextCursor, err := h.redis.Scan(c.Request.Context(), cursor, "price:all:*", 100).Result()
		if err != nil {
			Error(c, http.StatusInternalServerError, err.Error())
			return
		}
		for _, key := range keys {
			symbol := key[len("price:all:"):]
			members, _ := h.redis.ZRangeWithScores(c.Request.Context(), key, 0, -1).Result()
			exchanges := make([]gin.H, 0, len(members))
			for _, m := range members {
				exchanges = append(exchanges, gin.H{
					"exchange": m.Member, "price": fmt.Sprintf("%g", m.Score),
				})
			}
			allPrices = append(allPrices, gin.H{"symbol": symbol, "prices": exchanges})
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	OK(c, allPrices)
}
