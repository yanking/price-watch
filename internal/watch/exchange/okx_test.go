package exchange

import (
	"encoding/json"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/yanking/price-watch/internal/watch/event"
)

func TestParseOKXTickerMessage(t *testing.T) {
	msg := []byte(`{"arg":{"channel":"tickers","instId":"BTC-USDT"},"data":[{"instId":"BTC-USDT","last":"67000.5","vol24h":"1500.5","ts":"1711360000000"}]}`)

	a := &OKXAdapter{}
	result, err := a.parseOKXMessage(msg)
	if err != nil {
		t.Fatalf("parseOKXMessage: %v", err)
	}

	tick, ok := result.(event.Event[event.TickData])
	if !ok {
		t.Fatalf("expected Event[TickData], got %T", result)
	}

	if tick.Source != "okx" {
		t.Errorf("Source = %q, want %q", tick.Source, "okx")
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

func TestParseOKXNonTickerMessage(t *testing.T) {
	// Message with wrong channel
	msg := []byte(`{"arg":{"channel":"books","instId":"BTC-USDT"},"data":[]}`)
	a := &OKXAdapter{}
	_, err := a.parseOKXMessage(msg)
	if err == nil {
		t.Error("expected error for non-tickers channel message")
	}
}

func TestParseOKXEmptyData(t *testing.T) {
	msg := []byte(`{"arg":{"channel":"tickers","instId":"BTC-USDT"},"data":[]}`)
	a := &OKXAdapter{}
	_, err := a.parseOKXMessage(msg)
	if err == nil {
		t.Error("expected error for empty data array")
	}
}

func TestBuildOKXSubscribeMsg(t *testing.T) {
	msg, err := buildOKXSubscribeMsg([]string{"BTCUSDT", "ETHUSDT"})
	if err != nil {
		t.Fatalf("buildOKXSubscribeMsg: %v", err)
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

	arg0 := args[0].(map[string]any)
	if arg0["channel"] != "tickers" {
		t.Errorf("args[0].channel = %v, want tickers", arg0["channel"])
	}
	if arg0["instId"] != "BTC-USDT" {
		t.Errorf("args[0].instId = %v, want BTC-USDT", arg0["instId"])
	}

	arg1 := args[1].(map[string]any)
	if arg1["instId"] != "ETH-USDT" {
		t.Errorf("args[1].instId = %v, want ETH-USDT", arg1["instId"])
	}
}

func TestBuildOKXUnsubscribeMsg(t *testing.T) {
	msg, err := buildOKXUnsubscribeMsg([]string{"BTCUSDT"})
	if err != nil {
		t.Fatalf("buildOKXUnsubscribeMsg: %v", err)
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
	arg0 := args[0].(map[string]any)
	if arg0["instId"] != "BTC-USDT" {
		t.Errorf("args[0].instId = %v, want BTC-USDT", arg0["instId"])
	}
}

func TestNormalizeOKXSymbol(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"BTC-USDT", "BTCUSDT"},
		{"ETH-USDT", "ETHUSDT"},
		{"ETH-BTC", "ETHBTC"},
		{"BTCUSDT", "BTCUSDT"}, // already normalized
	}
	for _, tc := range cases {
		got := normalizeOKXSymbol(tc.input)
		if got != tc.want {
			t.Errorf("normalizeOKXSymbol(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestDenormalizeOKXSymbol(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"BTCUSDT", "BTC-USDT"},
		{"ETHUSDT", "ETH-USDT"},
		{"ETHBTC", "ETH-BTC"},
		{"BTCUSDC", "BTC-USDC"},
		{"ETHUSDC", "ETH-USDC"},
		{"SOLUSDT", "SOL-USDT"},
	}
	for _, tc := range cases {
		got := denormalizeOKXSymbol(tc.input)
		if got != tc.want {
			t.Errorf("denormalizeOKXSymbol(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
