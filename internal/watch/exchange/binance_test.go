package exchange

import (
	"encoding/json"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/yanking/price-watch/internal/watch/event"
)

func TestParseBinanceTradeMessage(t *testing.T) {
	msg := []byte(`{"e":"trade","E":1711360000000,"s":"BTCUSDT","t":123456,"p":"67000.50","q":"1.500","T":1711360000000,"m":true}`)

	a := &BinanceAdapter{}
	result, err := a.parseMessage(msg)
	if err != nil {
		t.Fatalf("parseMessage: %v", err)
	}

	tick, ok := result.(event.Event[event.TickData])
	if !ok {
		t.Fatalf("expected Event[TickData], got %T", result)
	}

	if tick.Source != "binance" {
		t.Errorf("Source = %q, want %q", tick.Source, "binance")
	}
	if tick.Subject != "BTCUSDT" {
		t.Errorf("Subject = %q, want %q", tick.Subject, "BTCUSDT")
	}
	if !tick.Data.Price.Equal(decimal.NewFromFloat(67000.50)) {
		t.Errorf("Price = %s, want 67000.50", tick.Data.Price)
	}
	if !tick.Data.Volume.Equal(decimal.NewFromFloat(1.5)) {
		t.Errorf("Volume = %s, want 1.5", tick.Data.Volume)
	}
}

func TestParseBinanceNonTradeMessage(t *testing.T) {
	msg := []byte(`{"result":null,"id":1}`)

	a := &BinanceAdapter{}
	_, err := a.parseMessage(msg)
	if err == nil {
		t.Error("expected error for non-trade message")
	}
}

func TestBuildBinanceSubscribeMsg(t *testing.T) {
	msg, err := buildBinanceSubscribeMsg([]string{"BTCUSDT", "ETHUSDT"})
	if err != nil {
		t.Fatalf("buildBinanceSubscribeMsg: %v", err)
	}

	var result map[string]any
	json.Unmarshal(msg, &result)

	if result["method"] != "SUBSCRIBE" {
		t.Errorf("method = %v, want SUBSCRIBE", result["method"])
	}

	params := result["params"].([]any)
	if len(params) != 2 {
		t.Fatalf("params length = %d, want 2", len(params))
	}
	if params[0] != "btcusdt@trade" {
		t.Errorf("params[0] = %v, want btcusdt@trade", params[0])
	}
	if params[1] != "ethusdt@trade" {
		t.Errorf("params[1] = %v, want ethusdt@trade", params[1])
	}
}

func TestBuildBinanceUnsubscribeMsg(t *testing.T) {
	msg, err := buildBinanceUnsubscribeMsg([]string{"BTCUSDT"})
	if err != nil {
		t.Fatalf("buildBinanceUnsubscribeMsg: %v", err)
	}

	var result map[string]any
	json.Unmarshal(msg, &result)

	if result["method"] != "UNSUBSCRIBE" {
		t.Errorf("method = %v, want UNSUBSCRIBE", result["method"])
	}
}
