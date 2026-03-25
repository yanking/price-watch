package event

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestEvent_BuildSubject(t *testing.T) {
	e := Event[TickData]{
		Type:    "price.tick",
		Source:  "binance",
		Subject: "BTCUSDT",
	}
	got := e.BuildSubject()
	want := "price.tick.binance.BTCUSDT"
	if got != want {
		t.Errorf("BuildSubject() = %q, want %q", got, want)
	}
}

func TestMarshalUnmarshal_TickData(t *testing.T) {
	original := Event[TickData]{
		ID:        "01ARZ3NDEKTSV4RRFFQ69G5FAV",
		Type:      "price.tick",
		Source:    "binance",
		Subject:   "BTCUSDT",
		Timestamp: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		Data: TickData{
			Price:  decimal.NewFromFloat(67000.50),
			Volume: decimal.NewFromFloat(1.5),
		},
	}

	data, err := Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	eventType, raw, err := Unmarshal(data)
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if eventType != "price.tick" {
		t.Errorf("eventType = %q, want %q", eventType, "price.tick")
	}
	if raw.Source != "binance" {
		t.Errorf("Source = %q, want %q", raw.Source, "binance")
	}

	tick, err := UnmarshalData[TickData](raw.Data)
	if err != nil {
		t.Fatalf("UnmarshalData: %v", err)
	}
	if !tick.Price.Equal(original.Data.Price) {
		t.Errorf("Price = %s, want %s", tick.Price, original.Data.Price)
	}
}

func TestMarshalUnmarshal_KlineData(t *testing.T) {
	original := Event[KlineData]{
		ID:        "01ARZ3NDEKTSV4RRFFQ69G5FAV",
		Type:      "price.kline",
		Source:    "okx",
		Subject:   "ETHUSDT",
		Timestamp: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		Data: KlineData{
			Interval: "1m",
			Open:     decimal.NewFromFloat(3500.00),
			High:     decimal.NewFromFloat(3550.00),
			Low:      decimal.NewFromFloat(3490.00),
			Close:    decimal.NewFromFloat(3520.00),
			Volume:   decimal.NewFromFloat(120.5),
		},
	}

	data, err := Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	eventType, raw, err := Unmarshal(data)
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if eventType != "price.kline" {
		t.Errorf("eventType = %q, want %q", eventType, "price.kline")
	}

	kline, err := UnmarshalData[KlineData](raw.Data)
	if err != nil {
		t.Fatalf("UnmarshalData: %v", err)
	}
	if kline.Interval != "1m" {
		t.Errorf("Interval = %q, want %q", kline.Interval, "1m")
	}
	if !kline.Close.Equal(original.Data.Close) {
		t.Errorf("Close = %s, want %s", kline.Close, original.Data.Close)
	}
}

func TestNewTickEvent(t *testing.T) {
	e := NewTickEvent("binance", "BTCUSDT", TickData{
		Price:  decimal.NewFromFloat(67000.50),
		Volume: decimal.NewFromFloat(1.5),
	})
	if e.ID == "" {
		t.Error("ID should not be empty")
	}
	if e.Type != TypeTick {
		t.Errorf("Type = %q, want %q", e.Type, TypeTick)
	}
	if e.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}
}

func TestNewKlineEvent(t *testing.T) {
	e := NewKlineEvent("okx", "ETHUSDT", KlineData{
		Interval: "5m",
		Open:     decimal.NewFromFloat(3500),
		High:     decimal.NewFromFloat(3550),
		Low:      decimal.NewFromFloat(3490),
		Close:    decimal.NewFromFloat(3520),
		Volume:   decimal.NewFromFloat(120.5),
	})
	if e.ID == "" {
		t.Error("ID should not be empty")
	}
	if e.Type != TypeKline {
		t.Errorf("Type = %q, want %q", e.Type, TypeKline)
	}
	if e.Source != "okx" {
		t.Errorf("Source = %q, want %q", e.Source, "okx")
	}
}
