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

type GateAdapter struct {
	*BaseAdapter
	restURL string
	limiter *rate.Limiter
}

func NewGateAdapter(bus eventbus.Bus, logger *slog.Logger, cfg ExchangeConfig) *GateAdapter {
	a := &GateAdapter{
		restURL: cfg.RestURL,
		limiter: rate.NewLimiter(rate.Every(70*time.Millisecond), 5), // ~15 req/s for 900 req/min
	}
	a.BaseAdapter = NewBaseAdapter(
		"gateio", bus, logger, cfg,
		a.parseGateMessage,
		buildGateSubscribeMsg,
		buildGateUnsubscribeMsg,
		buildGatePingMsg,
	)
	return a
}

func (a *GateAdapter) FetchKlines(ctx context.Context, symbol, interval string, start, end time.Time) ([]event.KlineData, error) {
	if err := a.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit: %w", err)
	}

	pair := denormalizeGateSymbol(symbol)
	url := fmt.Sprintf("%s/api/v4/spot/candlesticks?currency_pair=%s&interval=%s&from=%d&to=%d&limit=1000",
		a.restURL, pair, interval, start.Unix(), end.Unix())

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
		return nil, fmt.Errorf("gateio klines HTTP %d: %s", resp.StatusCode, string(body))
	}

	// Gate.io returns [][]string (or mixed array), each row:
	// [0]=ts(unix string), [1]=volume, [2]=close, [3]=high, [4]=low, [5]=open
	var raw [][]json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode klines: %w", err)
	}

	klines := make([]event.KlineData, 0, len(raw))
	for _, row := range raw {
		if len(row) < 6 {
			continue
		}
		// Gate.io order: [0]=ts, [1]=volume, [2]=close, [3]=high, [4]=low, [5]=open
		var volStr, closeStr, highStr, lowStr, openStr string
		json.Unmarshal(row[1], &volStr)
		json.Unmarshal(row[2], &closeStr)
		json.Unmarshal(row[3], &highStr)
		json.Unmarshal(row[4], &lowStr)
		json.Unmarshal(row[5], &openStr)

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

func (a *GateAdapter) parseGateMessage(msg []byte) (any, error) {
	var raw struct {
		Channel string `json:"channel"`
		Event   string `json:"event"`
		Result  struct {
			CurrencyPair string `json:"currency_pair"`
			Last         string `json:"last"`
			BaseVolume   string `json:"base_volume"`
		} `json:"result"`
	}
	if err := json.Unmarshal(msg, &raw); err != nil {
		return nil, err
	}

	if raw.Channel != "spot.tickers" || raw.Event != "update" {
		return nil, fmt.Errorf("not a spot.tickers update message")
	}

	symbol := normalizeGateSymbol(raw.Result.CurrencyPair)

	price, err := decimal.NewFromString(raw.Result.Last)
	if err != nil {
		return nil, fmt.Errorf("parse price: %w", err)
	}
	vol, _ := decimal.NewFromString(raw.Result.BaseVolume)

	return event.NewTickEvent("gateio", symbol, event.TickData{
		Price:  price,
		Volume: vol,
	}), nil
}

func buildGateSubscribeMsg(symbols []string) ([]byte, error) {
	payload := make([]string, len(symbols))
	for i, s := range symbols {
		payload[i] = denormalizeGateSymbol(s)
	}
	return json.Marshal(map[string]any{
		"time":    time.Now().Unix(),
		"channel": "spot.tickers",
		"event":   "subscribe",
		"payload": payload,
	})
}

func buildGateUnsubscribeMsg(symbols []string) ([]byte, error) {
	payload := make([]string, len(symbols))
	for i, s := range symbols {
		payload[i] = denormalizeGateSymbol(s)
	}
	return json.Marshal(map[string]any{
		"time":    time.Now().Unix(),
		"channel": "spot.tickers",
		"event":   "unsubscribe",
		"payload": payload,
	})
}

func buildGatePingMsg() []byte {
	b, _ := json.Marshal(map[string]any{
		"time":    time.Now().Unix(),
		"channel": "spot.ping",
	})
	return b
}

// normalizeGateSymbol converts Gate.io format (BTC_USDT) to canonical format (BTCUSDT).
func normalizeGateSymbol(s string) string {
	return strings.ReplaceAll(s, "_", "")
}

// denormalizeGateSymbol converts canonical format (BTCUSDT) to Gate.io format (BTC_USDT).
// It recognises common quote currencies and inserts the underscore before them.
func denormalizeGateSymbol(s string) string {
	quotes := []string{"USDT", "USDC", "BTC", "ETH"}
	upper := strings.ToUpper(s)
	for _, q := range quotes {
		if strings.HasSuffix(upper, q) && len(s) > len(q) {
			base := s[:len(s)-len(q)]
			return base + "_" + q
		}
	}
	return s
}
