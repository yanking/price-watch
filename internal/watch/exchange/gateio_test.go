package exchange

import (
	"encoding/json"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/yanking/price-watch/internal/watch/event"
)

func TestParseGateTickerMessage(t *testing.T) {
	msg := []byte(`{"time":1711360000,"channel":"spot.tickers","event":"update","result":{"currency_pair":"BTC_USDT","last":"67000.5","base_volume":"1500.5"}}`)

	a := &GateAdapter{}
	result, err := a.parseGateMessage(msg)
	if err != nil {
		t.Fatalf("parseGateMessage: %v", err)
	}

	tick, ok := result.(event.Event[event.TickData])
	if !ok {
		t.Fatalf("expected Event[TickData], got %T", result)
	}

	if tick.Source != "gateio" {
		t.Errorf("Source = %q, want %q", tick.Source, "gateio")
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

func TestParseGateNonUpdateMessage(t *testing.T) {
	// Wrong event type (subscribe ack, not update)
	msg := []byte(`{"time":1711360000,"channel":"spot.tickers","event":"subscribe","result":{"status":"success"}}`)
	a := &GateAdapter{}
	_, err := a.parseGateMessage(msg)
	if err == nil {
		t.Error("expected error for non-update event message")
	}
}

func TestParseGateWrongChannelMessage(t *testing.T) {
	// Wrong channel
	msg := []byte(`{"time":1711360000,"channel":"spot.trades","event":"update","result":{}}`)
	a := &GateAdapter{}
	_, err := a.parseGateMessage(msg)
	if err == nil {
		t.Error("expected error for non spot.tickers channel message")
	}
}

func TestParseGateInvalidJSON(t *testing.T) {
	msg := []byte(`not valid json`)
	a := &GateAdapter{}
	_, err := a.parseGateMessage(msg)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestBuildGateSubscribeMsg(t *testing.T) {
	msg, err := buildGateSubscribeMsg([]string{"BTCUSDT", "ETHUSDT"})
	if err != nil {
		t.Fatalf("buildGateSubscribeMsg: %v", err)
	}

	var result map[string]any
	json.Unmarshal(msg, &result)

	if result["channel"] != "spot.tickers" {
		t.Errorf("channel = %v, want spot.tickers", result["channel"])
	}
	if result["event"] != "subscribe" {
		t.Errorf("event = %v, want subscribe", result["event"])
	}
	if _, ok := result["time"]; !ok {
		t.Error("expected time field in subscribe message")
	}

	payload := result["payload"].([]any)
	if len(payload) != 2 {
		t.Fatalf("payload length = %d, want 2", len(payload))
	}
	if payload[0] != "BTC_USDT" {
		t.Errorf("payload[0] = %v, want BTC_USDT", payload[0])
	}
	if payload[1] != "ETH_USDT" {
		t.Errorf("payload[1] = %v, want ETH_USDT", payload[1])
	}
}

func TestBuildGateUnsubscribeMsg(t *testing.T) {
	msg, err := buildGateUnsubscribeMsg([]string{"BTCUSDT"})
	if err != nil {
		t.Fatalf("buildGateUnsubscribeMsg: %v", err)
	}

	var result map[string]any
	json.Unmarshal(msg, &result)

	if result["channel"] != "spot.tickers" {
		t.Errorf("channel = %v, want spot.tickers", result["channel"])
	}
	if result["event"] != "unsubscribe" {
		t.Errorf("event = %v, want unsubscribe", result["event"])
	}
	if _, ok := result["time"]; !ok {
		t.Error("expected time field in unsubscribe message")
	}

	payload := result["payload"].([]any)
	if len(payload) != 1 {
		t.Fatalf("payload length = %d, want 1", len(payload))
	}
	if payload[0] != "BTC_USDT" {
		t.Errorf("payload[0] = %v, want BTC_USDT", payload[0])
	}
}

func TestNormalizeGateSymbol(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"BTC_USDT", "BTCUSDT"},
		{"ETH_USDT", "ETHUSDT"},
		{"ETH_BTC", "ETHBTC"},
		{"BTCUSDT", "BTCUSDT"}, // already normalized (no underscore)
		{"SOL_USDT", "SOLUSDT"},
	}
	for _, tc := range cases {
		got := normalizeGateSymbol(tc.input)
		if got != tc.want {
			t.Errorf("normalizeGateSymbol(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestDenormalizeGateSymbol(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"BTCUSDT", "BTC_USDT"},
		{"ETHUSDT", "ETH_USDT"},
		{"ETHBTC", "ETH_BTC"},
		{"BTCUSDC", "BTC_USDC"},
		{"ETHUSDC", "ETH_USDC"},
		{"SOLUSDT", "SOL_USDT"},
	}
	for _, tc := range cases {
		got := denormalizeGateSymbol(tc.input)
		if got != tc.want {
			t.Errorf("denormalizeGateSymbol(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestBuildGatePingMsg(t *testing.T) {
	msg := buildGatePingMsg()
	if msg == nil {
		t.Fatal("expected non-nil ping message")
	}

	var result map[string]any
	if err := json.Unmarshal(msg, &result); err != nil {
		t.Fatalf("ping message is not valid JSON: %v", err)
	}
	if result["channel"] != "spot.ping" {
		t.Errorf("channel = %v, want spot.ping", result["channel"])
	}
	if _, ok := result["time"]; !ok {
		t.Error("expected time field in ping message")
	}
}
