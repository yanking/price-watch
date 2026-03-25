package exchange

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/yanking/price-watch/internal/watch/event"
	"github.com/yanking/price-watch/pkg/eventbus"
	"golang.org/x/time/rate"
)

type BinanceAdapter struct {
	*BaseAdapter
	restURL string
	limiter *rate.Limiter
}

func NewBinanceAdapter(bus eventbus.Bus, logger *slog.Logger, cfg ExchangeConfig) *BinanceAdapter {
	a := &BinanceAdapter{
		restURL: cfg.RestURL,
		limiter: rate.NewLimiter(rate.Every(50*time.Millisecond), 5), // ~20 req/s
	}
	a.BaseAdapter = NewBaseAdapter(
		"binance", bus, logger, cfg,
		a.parseMessage,
		buildBinanceSubscribeMsg,
		buildBinanceUnsubscribeMsg,
		func() []byte { return nil }, // use WebSocket ping frame
	)
	return a
}

func (a *BinanceAdapter) FetchKlines(ctx context.Context, symbol, interval string, start, end time.Time) ([]event.KlineData, error) {
	if err := a.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit: %w", err)
	}

	url := fmt.Sprintf("%s/api/v3/klines?symbol=%s&interval=%s&startTime=%d&endTime=%d&limit=1000",
		a.restURL, symbol, interval, start.UnixMilli(), end.UnixMilli())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch klines: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("binance klines HTTP %d: %s", resp.StatusCode, string(body))
	}

	var raw [][]json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode klines: %w", err)
	}

	klines := make([]event.KlineData, 0, len(raw))
	for _, row := range raw {
		if len(row) < 6 {
			continue
		}
		var openStr, highStr, lowStr, closeStr, volStr string
		json.Unmarshal(row[1], &openStr)
		json.Unmarshal(row[2], &highStr)
		json.Unmarshal(row[3], &lowStr)
		json.Unmarshal(row[4], &closeStr)
		json.Unmarshal(row[5], &volStr)

		klines = append(klines, event.KlineData{
			Interval: interval,
			Open:     decimalFromString(openStr),
			High:     decimalFromString(highStr),
			Low:      decimalFromString(lowStr),
			Close:    decimalFromString(closeStr),
			Volume:   decimalFromString(volStr),
		})
	}

	return klines, nil
}

func (a *BinanceAdapter) parseMessage(msg []byte) (any, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(msg, &raw); err != nil {
		return nil, err
	}

	// Check if this is a trade event
	var eventType string
	if e, ok := raw["e"]; ok {
		json.Unmarshal(e, &eventType)
	}
	if eventType != "trade" {
		return nil, fmt.Errorf("not a trade event")
	}

	var symbol, priceStr, qtyStr string
	json.Unmarshal(raw["s"], &symbol)
	json.Unmarshal(raw["p"], &priceStr)
	json.Unmarshal(raw["q"], &qtyStr)

	return event.NewTickEvent("binance", symbol, event.TickData{
		Price:  decimalFromString(priceStr),
		Volume: decimalFromString(qtyStr),
	}), nil
}

func buildBinanceSubscribeMsg(symbols []string) ([]byte, error) {
	params := make([]string, len(symbols))
	for i, s := range symbols {
		params[i] = strings.ToLower(s) + "@trade"
	}
	return json.Marshal(map[string]any{
		"method": "SUBSCRIBE",
		"params": params,
		"id":     1,
	})
}

func buildBinanceUnsubscribeMsg(symbols []string) ([]byte, error) {
	params := make([]string, len(symbols))
	for i, s := range symbols {
		params[i] = strings.ToLower(s) + "@trade"
	}
	return json.Marshal(map[string]any{
		"method": "UNSUBSCRIBE",
		"params": params,
		"id":     1,
	})
}

func decimalFromString(s string) decimal.Decimal {
	d, _ := decimal.NewFromString(s)
	return d
}
