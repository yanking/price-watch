package consumer

import (
	"log/slog"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/yanking/price-watch/internal/watch/event"
)

func TestStorageSubscriber_HandleTick_CreatesPoint(t *testing.T) {
	// Create subscriber without real InfluxDB — just test point buffering
	s := &StorageSubscriber{
		logger:    slog.Default(),
		batchSize: 100,
		stopCh:    make(chan struct{}),
		stopped:   make(chan struct{}),
	}
	go func() { <-s.stopCh; close(s.stopped) }()

	tick := event.Event[event.TickData]{
		ID:        "test-id",
		Type:      event.TypeTick,
		Source:    "binance",
		Subject:   "BTCUSDT",
		Timestamp: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		Data: event.TickData{
			Price:  decimal.NewFromFloat(67000.50),
			Volume: decimal.NewFromFloat(1.5),
		},
	}

	err := s.handleTick(tick)
	if err != nil {
		t.Fatalf("handleTick: %v", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.tickBuf) != 1 {
		t.Fatalf("tickBuf length = %d, want 1", len(s.tickBuf))
	}
}

func TestStorageSubscriber_HandleKline_CreatesPoint(t *testing.T) {
	s := &StorageSubscriber{
		logger:    slog.Default(),
		batchSize: 100,
		stopCh:    make(chan struct{}),
		stopped:   make(chan struct{}),
	}
	go func() { <-s.stopCh; close(s.stopped) }()

	kline := event.Event[event.KlineData]{
		ID:        "test-id",
		Type:      event.TypeKline,
		Source:    "okx",
		Subject:   "ETHUSDT",
		Timestamp: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		Data: event.KlineData{
			Interval: "1m",
			Open:     decimal.NewFromFloat(3500),
			High:     decimal.NewFromFloat(3550),
			Low:      decimal.NewFromFloat(3490),
			Close:    decimal.NewFromFloat(3520),
			Volume:   decimal.NewFromFloat(120.5),
		},
	}

	err := s.handleKline(kline)
	if err != nil {
		t.Fatalf("handleKline: %v", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.klineBuf) != 1 {
		t.Fatalf("klineBuf length = %d, want 1", len(s.klineBuf))
	}
}

func TestStorageSubscriber_BatchSize(t *testing.T) {
	// With batchSize=2, after 2 ticks the buffer should trigger flush
	// Since we don't have real InfluxDB, the flush will fail silently (logger warns)
	s := &StorageSubscriber{
		logger:    slog.Default(),
		batchSize: 2,
		stopCh:    make(chan struct{}),
		stopped:   make(chan struct{}),
	}
	go func() { <-s.stopCh; close(s.stopped) }()

	for i := 0; i < 3; i++ {
		tick := event.Event[event.TickData]{
			ID:        "test",
			Type:      event.TypeTick,
			Source:    "binance",
			Subject:   "BTCUSDT",
			Timestamp: time.Now(),
			Data: event.TickData{
				Price:  decimal.NewFromFloat(67000),
				Volume: decimal.NewFromFloat(1),
			},
		}
		s.handleTick(tick)
	}

	// After 3 ticks with batchSize=2, one flush should have happened (clearing 2),
	// leaving 1 in the buffer
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.tickBuf) != 1 {
		t.Errorf("tickBuf length = %d, want 1 (after batch flush)", len(s.tickBuf))
	}
}
