package consumer

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"github.com/yanking/price-watch/internal/watch/event"
	"github.com/yanking/price-watch/pkg/eventbus"
)

func TestCacheSubscriber_HandleTick(t *testing.T) {
	mr := miniredis.RunT(t)
	rds := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	bus, _ := eventbus.New(eventbus.Config{Driver: "memory", Memory: eventbus.MemoryConfig{BufferSize: 64}}, slog.Default())
	defer bus.Close()

	sub := NewCacheSubscriber(rds, bus, slog.Default())
	if err := sub.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer sub.Stop()

	// Publish a tick event
	tick := event.NewTickEvent("binance", "BTCUSDT", event.TickData{
		Price:  decimal.NewFromFloat(67000.50),
		Volume: decimal.NewFromFloat(1.5),
	})
	data, _ := event.Marshal(tick)
	bus.Publish(context.Background(), tick.BuildSubject(), data)

	// Wait for async processing
	time.Sleep(100 * time.Millisecond)

	// Verify Hash
	hashKey := "price:binance:BTCUSDT"
	price := mr.HGet(hashKey, "price")
	if price != "67000.5" {
		t.Errorf("price = %q, want %q", price, "67000.5")
	}

	volume := mr.HGet(hashKey, "volume")
	if volume != "1.5" {
		t.Errorf("volume = %q, want %q", volume, "1.5")
	}

	// Verify Sorted Set
	zsetKey := "price:all:BTCUSDT"
	members, err := mr.ZMembers(zsetKey)
	if err != nil {
		t.Fatalf("ZMembers: %v", err)
	}
	if len(members) != 1 || members[0] != "binance" {
		t.Errorf("ZMembers = %v, want [binance]", members)
	}

	score, err := mr.ZScore(zsetKey, "binance")
	if err != nil {
		t.Fatalf("ZScore: %v", err)
	}
	if score != 67000.5 {
		t.Errorf("ZScore = %f, want 67000.5", score)
	}
}

func TestCacheSubscriber_MultipleExchanges(t *testing.T) {
	mr := miniredis.RunT(t)
	rds := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	bus, _ := eventbus.New(eventbus.Config{Driver: "memory", Memory: eventbus.MemoryConfig{BufferSize: 64}}, slog.Default())
	defer bus.Close()

	sub := NewCacheSubscriber(rds, bus, slog.Default())
	sub.Start()
	defer sub.Stop()

	// Publish from two exchanges
	for _, exchange := range []string{"binance", "okx"} {
		tick := event.NewTickEvent(exchange, "BTCUSDT", event.TickData{
			Price:  decimal.NewFromFloat(67000),
			Volume: decimal.NewFromFloat(1),
		})
		data, _ := event.Marshal(tick)
		bus.Publish(context.Background(), tick.BuildSubject(), data)
	}

	time.Sleep(100 * time.Millisecond)

	// Verify both exchanges in sorted set
	members, _ := mr.ZMembers("price:all:BTCUSDT")
	if len(members) != 2 {
		t.Errorf("ZMembers count = %d, want 2", len(members))
	}
}
