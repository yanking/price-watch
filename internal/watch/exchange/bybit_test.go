package exchange

import (
	"encoding/json"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/yanking/price-watch/internal/watch/event"
)

func TestParseBybitTickerMessage(t *testing.T) {
	msg := []byte(`{"topic":"tickers.BTCUSDT","type":"snapshot","data":{"symbol":"BTCUSDT","lastPrice":"67000.5","volume24h":"1500.5"}}`)

	a := &BybitAdapter{}
	result, err := a.parseBybitMessage(msg)
	if err != nil {
		t.Fatalf("parseBybitMessage: %v", err)
	}

	tick, ok := result.(event.Event[event.TickData])
	if !ok {
		t.Fatalf("expected Event[TickData], got %T", result)
	}

	if tick.Source != "bybit" {
		t.Errorf("Source = %q, want %q", tick.Source, "bybit")
	}
	if tick.Subject != "BTCUSDT" {
		t.Errorf("Subject = %q, want %q", tick.Subject, "BTCUSDT")
	}
	want := decimal.NewFromFloat(67000.5)
	if !tick.Data.Price.Equal(want) {
		t.Errorf("Price = %s, want %s", tick.Data.Price, want)
	}
	wantVol := decimal.NewFromFloat(1500.5)
	if !tick.Data.Volume.Equal(wantVol) {
		t.Errorf("Volume = %s, want %s", tick.Data.Volume, wantVol)
	}
}

func TestParseBybitNonTickerMessage(t *testing.T) {
	msg := []byte(`{"topic":"orderbook.BTCUSDT","type":"snapshot","data":{}}`)

	a := &BybitAdapter{}
	_, err := a.parseBybitMessage(msg)
	if err == nil {
		t.Error("expected error for non-tickers topic message")
	}
}

func TestParseBybitInvalidJSON(t *testing.T) {
	msg := []byte(`not valid json`)

	a := &BybitAdapter{}
	_, err := a.parseBybitMessage(msg)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestBuildBybitSubscribeMsg(t *testing.T) {
	msg, err := buildBybitSubscribeMsg([]string{"BTCUSDT", "ETHUSDT"})
	if err != nil {
		t.Fatalf("buildBybitSubscribeMsg: %v", err)
	}

	var result map[string]any
	json.Unmarshal(msg, &result)

	if result["op"] != "subscribe" {
		t.Errorf("op = %v, want subscribe", result["op"])
	}

	args := result["args"].([]any)
	if len(args) != 2 {
		t.Fatalf("args length = %d, want 2", len(args))
	}
	if args[0] != "tickers.BTCUSDT" {
		t.Errorf("args[0] = %v, want tickers.BTCUSDT", args[0])
	}
	if args[1] != "tickers.ETHUSDT" {
		t.Errorf("args[1] = %v, want tickers.ETHUSDT", args[1])
	}
}

func TestBuildBybitUnsubscribeMsg(t *testing.T) {
	msg, err := buildBybitUnsubscribeMsg([]string{"BTCUSDT"})
	if err != nil {
		t.Fatalf("buildBybitUnsubscribeMsg: %v", err)
	}

	var result map[string]any
	json.Unmarshal(msg, &result)

	if result["op"] != "unsubscribe" {
		t.Errorf("op = %v, want unsubscribe", result["op"])
	}

	args := result["args"].([]any)
	if len(args) != 1 {
		t.Fatalf("args length = %d, want 1", len(args))
	}
	if args[0] != "tickers.BTCUSDT" {
		t.Errorf("args[0] = %v, want tickers.BTCUSDT", args[0])
	}
}

func TestMapBybitInterval(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"1m", "1"},
		{"5m", "5"},
		{"15m", "15"},
		{"1h", "60"},
		{"4h", "240"},
		{"1d", "D"},
		{"unknown", "unknown"}, // passthrough for unrecognized intervals
	}
	for _, tc := range cases {
		got := mapBybitInterval(tc.input)
		if got != tc.want {
			t.Errorf("mapBybitInterval(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
