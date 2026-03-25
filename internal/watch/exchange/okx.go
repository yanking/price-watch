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

type OKXAdapter struct {
	*BaseAdapter
	restURL string
	limiter *rate.Limiter
}

func NewOKXAdapter(bus eventbus.Bus, logger *slog.Logger, cfg ExchangeConfig) *OKXAdapter {
	a := &OKXAdapter{
		restURL: cfg.RestURL,
		limiter: rate.NewLimiter(rate.Every(100*time.Millisecond), 2), // ~10 req/s
	}
	a.BaseAdapter = NewBaseAdapter(
		"okx", bus, logger, cfg,
		a.parseOKXMessage,
		buildOKXSubscribeMsg,
		buildOKXUnsubscribeMsg,
		func() []byte { return []byte("ping") },
	)
	return a
}

func (a *OKXAdapter) FetchKlines(ctx context.Context, symbol, interval string, start, end time.Time) ([]event.KlineData, error) {
	if err := a.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit: %w", err)
	}

	instID := denormalizeOKXSymbol(symbol)
	url := fmt.Sprintf("%s/api/v5/market/candles?instId=%s&bar=%s&limit=300",
		a.restURL, instID, interval)

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
		return nil, fmt.Errorf("okx klines HTTP %d: %s", resp.StatusCode, string(body))
	}

	var raw struct {
		Code string              `json:"code"`
		Data [][]json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode klines: %w", err)
	}
	if raw.Code != "0" {
		return nil, fmt.Errorf("okx klines error code: %s", raw.Code)
	}

	klines := make([]event.KlineData, 0, len(raw.Data))
	for _, row := range raw.Data {
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

func (a *OKXAdapter) parseOKXMessage(msg []byte) (any, error) {
	var raw struct {
		Arg struct {
			Channel string `json:"channel"`
			InstID  string `json:"instId"`
		} `json:"arg"`
		Data []struct {
			InstID string `json:"instId"`
			Last   string `json:"last"`
			Vol24h string `json:"vol24h"`
			Ts     string `json:"ts"`
		} `json:"data"`
	}
	if err := json.Unmarshal(msg, &raw); err != nil {
		return nil, err
	}

	if raw.Arg.Channel != "tickers" || len(raw.Data) == 0 {
		return nil, fmt.Errorf("not a tickers channel message")
	}

	d := raw.Data[0]
	symbol := normalizeOKXSymbol(d.InstID)

	// Parse price as decimal
	price, err := decimal.NewFromString(d.Last)
	if err != nil {
		return nil, fmt.Errorf("parse price: %w", err)
	}
	vol, _ := decimal.NewFromString(d.Vol24h)

	return event.NewTickEvent("okx", symbol, event.TickData{
		Price:  price,
		Volume: vol,
	}), nil
}

func buildOKXSubscribeMsg(symbols []string) ([]byte, error) {
	type arg struct {
		Channel string `json:"channel"`
		InstID  string `json:"instId"`
	}
	args := make([]arg, len(symbols))
	for i, s := range symbols {
		args[i] = arg{Channel: "tickers", InstID: denormalizeOKXSymbol(s)}
	}
	return json.Marshal(map[string]any{
		"op":   "subscribe",
		"args": args,
	})
}

func buildOKXUnsubscribeMsg(symbols []string) ([]byte, error) {
	type arg struct {
		Channel string `json:"channel"`
		InstID  string `json:"instId"`
	}
	args := make([]arg, len(symbols))
	for i, s := range symbols {
		args[i] = arg{Channel: "tickers", InstID: denormalizeOKXSymbol(s)}
	}
	return json.Marshal(map[string]any{
		"op":   "unsubscribe",
		"args": args,
	})
}

// normalizeOKXSymbol converts OKX format (BTC-USDT) to canonical format (BTCUSDT).
func normalizeOKXSymbol(s string) string {
	return strings.ReplaceAll(s, "-", "")
}

// denormalizeOKXSymbol converts canonical format (BTCUSDT) to OKX format (BTC-USDT).
// It recognises common quote currencies and inserts the dash before them.
func denormalizeOKXSymbol(s string) string {
	quotes := []string{"USDT", "USDC", "BTC", "ETH"}
	upper := strings.ToUpper(s)
	for _, q := range quotes {
		if strings.HasSuffix(upper, q) && len(s) > len(q) {
			base := s[:len(s)-len(q)]
			return base + "-" + q
		}
	}
	return s
}
