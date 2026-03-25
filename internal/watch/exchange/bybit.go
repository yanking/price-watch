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

type BybitAdapter struct {
	*BaseAdapter
	restURL string
	limiter *rate.Limiter
}

func NewBybitAdapter(bus eventbus.Bus, logger *slog.Logger, cfg ExchangeConfig) *BybitAdapter {
	a := &BybitAdapter{
		restURL: cfg.RestURL,
		limiter: rate.NewLimiter(rate.Every(500*time.Millisecond), 2), // ~2 req/s, conservative for 120 req/min
	}
	a.BaseAdapter = NewBaseAdapter(
		"bybit", bus, logger, cfg,
		a.parseBybitMessage,
		buildBybitSubscribeMsg,
		buildBybitUnsubscribeMsg,
		func() []byte {
			b, _ := json.Marshal(map[string]string{"op": "ping"})
			return b
		},
	)
	return a
}

func (a *BybitAdapter) FetchKlines(ctx context.Context, symbol, interval string, start, end time.Time) ([]event.KlineData, error) {
	if err := a.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit: %w", err)
	}

	bybitInterval := mapBybitInterval(interval)
	url := fmt.Sprintf("%s/v5/market/kline?category=spot&symbol=%s&interval=%s&start=%d&end=%d&limit=200",
		a.restURL, symbol, bybitInterval, start.UnixMilli(), end.UnixMilli())

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
		return nil, fmt.Errorf("bybit klines HTTP %d: %s", resp.StatusCode, string(body))
	}

	var raw struct {
		RetCode int `json:"retCode"`
		Result  struct {
			List [][]json.RawMessage `json:"list"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode klines: %w", err)
	}
	if raw.RetCode != 0 {
		return nil, fmt.Errorf("bybit klines error retCode: %d", raw.RetCode)
	}

	klines := make([]event.KlineData, 0, len(raw.Result.List))
	for _, row := range raw.Result.List {
		// [0]=ts, [1]=open, [2]=high, [3]=low, [4]=close, [5]=volume
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

func (a *BybitAdapter) parseBybitMessage(msg []byte) (any, error) {
	var raw struct {
		Topic string `json:"topic"`
		Type  string `json:"type"`
		Data  struct {
			Symbol    string `json:"symbol"`
			LastPrice string `json:"lastPrice"`
			Volume24h string `json:"volume24h"`
		} `json:"data"`
	}
	if err := json.Unmarshal(msg, &raw); err != nil {
		return nil, err
	}

	// Only handle tickers.* snapshot messages
	if !strings.HasPrefix(raw.Topic, "tickers.") {
		return nil, fmt.Errorf("not a tickers topic message")
	}

	price, err := decimal.NewFromString(raw.Data.LastPrice)
	if err != nil {
		return nil, fmt.Errorf("parse price: %w", err)
	}
	vol, _ := decimal.NewFromString(raw.Data.Volume24h)

	return event.NewTickEvent("bybit", raw.Data.Symbol, event.TickData{
		Price:  price,
		Volume: vol,
	}), nil
}

func buildBybitSubscribeMsg(symbols []string) ([]byte, error) {
	args := make([]string, len(symbols))
	for i, s := range symbols {
		args[i] = "tickers." + s
	}
	return json.Marshal(map[string]any{
		"op":   "subscribe",
		"args": args,
	})
}

func buildBybitUnsubscribeMsg(symbols []string) ([]byte, error) {
	args := make([]string, len(symbols))
	for i, s := range symbols {
		args[i] = "tickers." + s
	}
	return json.Marshal(map[string]any{
		"op":   "unsubscribe",
		"args": args,
	})
}

// mapBybitInterval maps canonical interval strings to Bybit API interval values.
func mapBybitInterval(interval string) string {
	switch interval {
	case "1m":
		return "1"
	case "5m":
		return "5"
	case "15m":
		return "15"
	case "1h":
		return "60"
	case "4h":
		return "240"
	case "1d":
		return "D"
	default:
		return interval
	}
}
