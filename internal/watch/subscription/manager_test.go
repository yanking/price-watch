package subscription

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/yanking/price-watch/internal/watch/event"
	"github.com/yanking/price-watch/internal/watch/exchange"
)

// mockAdapter implements exchange.ExchangeAdapter for testing
type mockAdapter struct {
	name       string
	subscribed []string
	failOn     string // if set, Subscribe/Unsubscribe returns error for this symbol
}

func (m *mockAdapter) Name() string { return m.name }
func (m *mockAdapter) Subscribe(symbols []string) error {
	for _, s := range symbols {
		if s == m.failOn {
			return fmt.Errorf("mock subscribe error")
		}
	}
	m.subscribed = append(m.subscribed, symbols...)
	return nil
}
func (m *mockAdapter) Unsubscribe(symbols []string) error {
	if len(symbols) > 0 && symbols[0] == m.failOn {
		return fmt.Errorf("mock unsubscribe error")
	}
	return nil
}
func (m *mockAdapter) FetchKlines(ctx context.Context, symbol, interval string, start, end time.Time) ([]event.KlineData, error) {
	return nil, nil
}
func (m *mockAdapter) Start(ctx context.Context) error { return nil }
func (m *mockAdapter) Stop() error                     { return nil }

func setupTest(t *testing.T) (*Manager, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	rds := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	adapters := map[string]exchange.ExchangeAdapter{
		"binance": &mockAdapter{name: "binance"},
		"okx":     &mockAdapter{name: "okx"},
	}

	mgr := NewManager(rds, adapters, slog.Default())
	return mgr, mr
}

func TestManager_Add(t *testing.T) {
	mgr, mr := setupTest(t)
	ctx := context.Background()

	result, err := mgr.Add(ctx, SubscribeRequest{
		Exchanges: []string{"binance"},
		Symbols:   []string{"BTCUSDT", "ETHUSDT"},
	})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if len(result.Succeeded) != 1 || result.Succeeded[0] != "binance" {
		t.Errorf("Succeeded = %v, want [binance]", result.Succeeded)
	}

	// Verify Redis
	members, _ := mr.SMembers("sub:binance")
	if len(members) != 2 {
		t.Errorf("Redis members = %d, want 2", len(members))
	}
}

func TestManager_Add_AllExchanges(t *testing.T) {
	mgr, _ := setupTest(t)
	ctx := context.Background()

	result, err := mgr.Add(ctx, SubscribeRequest{
		Symbols: []string{"BTCUSDT"},
	})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if len(result.Succeeded) != 2 {
		t.Errorf("Succeeded count = %d, want 2", len(result.Succeeded))
	}
}

func TestManager_Remove(t *testing.T) {
	mgr, mr := setupTest(t)
	ctx := context.Background()

	// First add
	mgr.Add(ctx, SubscribeRequest{
		Exchanges: []string{"binance"},
		Symbols:   []string{"BTCUSDT", "ETHUSDT"},
	})

	// Then remove one
	result, err := mgr.Remove(ctx, UnsubscribeRequest{
		Exchanges: []string{"binance"},
		Symbols:   []string{"BTCUSDT"},
	})
	if err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if len(result.Succeeded) != 1 {
		t.Errorf("Succeeded = %v, want [binance]", result.Succeeded)
	}

	members, _ := mr.SMembers("sub:binance")
	if len(members) != 1 {
		t.Errorf("Redis members = %d, want 1", len(members))
	}
}

func TestManager_List(t *testing.T) {
	mgr, _ := setupTest(t)
	ctx := context.Background()

	mgr.Add(ctx, SubscribeRequest{
		Exchanges: []string{"binance"},
		Symbols:   []string{"BTCUSDT"},
	})

	infos, err := mgr.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	found := false
	for _, info := range infos {
		if info.Exchange == "binance" && len(info.Symbols) == 1 {
			found = true
		}
	}
	if !found {
		t.Errorf("expected binance with 1 symbol, got %v", infos)
	}
}

func TestManager_Restore(t *testing.T) {
	mgr, mr := setupTest(t)
	ctx := context.Background()

	// Pre-populate Redis
	mr.SAdd("sub:binance", "BTCUSDT", "ETHUSDT")

	err := mgr.Restore(ctx)
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}

	// Verify adapter was called
	adapter := mgr.adapters["binance"].(*mockAdapter)
	if len(adapter.subscribed) != 2 {
		t.Errorf("adapter subscribed = %d, want 2", len(adapter.subscribed))
	}
}

func TestManager_PartialFailure(t *testing.T) {
	mr := miniredis.RunT(t)
	rds := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	adapters := map[string]exchange.ExchangeAdapter{
		"binance": &mockAdapter{name: "binance"},
		"okx":     &mockAdapter{name: "okx", failOn: "BTCUSDT"},
	}
	mgr := NewManager(rds, adapters, slog.Default())
	ctx := context.Background()

	result, err := mgr.Add(ctx, SubscribeRequest{
		Symbols: []string{"BTCUSDT"},
	})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	// binance should succeed, okx should fail
	hasSucceeded := false
	for _, s := range result.Succeeded {
		if s == "binance" {
			hasSucceeded = true
		}
	}
	if !hasSucceeded {
		t.Error("expected binance in succeeded")
	}
	if _, ok := result.Failed["okx"]; !ok {
		t.Error("expected okx in failed")
	}
}
