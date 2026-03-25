package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func TestPriceHandler_GetBySymbol(t *testing.T) {
	mr := miniredis.RunT(t)
	rds := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	// Seed Redis with test data
	mr.ZAdd("price:all:BTCUSDT", 67000.5, "binance")
	mr.ZAdd("price:all:BTCUSDT", 67001.0, "okx")

	router := gin.New()
	h := NewPriceHandler(rds)
	router.GET("/api/v1/prices/:symbol", h.GetBySymbol)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/prices/BTCUSDT", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != 0 {
		t.Errorf("response code = %d, want 0", resp.Code)
	}

	// Verify prices are in response data
	data, ok := resp.Data.(map[string]any)
	if !ok {
		t.Fatal("response data is not a map")
	}
	prices, ok := data["prices"].([]any)
	if !ok {
		t.Fatal("prices is not an array")
	}
	if len(prices) != 2 {
		t.Errorf("prices count = %d, want 2", len(prices))
	}
}

func TestPriceHandler_GetBySymbol_Empty(t *testing.T) {
	mr := miniredis.RunT(t)
	rds := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	router := gin.New()
	h := NewPriceHandler(rds)
	router.GET("/api/v1/prices/:symbol", h.GetBySymbol)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/prices/UNKNOWN", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}
