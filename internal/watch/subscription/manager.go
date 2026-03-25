package subscription

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/redis/go-redis/v9"
	"github.com/yanking/price-watch/internal/watch/exchange"
)

type SubscribeRequest struct {
	Exchanges []string `json:"exchanges"`
	Symbols   []string `json:"symbols"`
}

type UnsubscribeRequest = SubscribeRequest

type SubscribeResult struct {
	Succeeded []string          `json:"succeeded"`
	Failed    map[string]string `json:"failed,omitempty"`
}

type SubscriptionInfo struct {
	Exchange string   `json:"exchange"`
	Symbols  []string `json:"symbols"`
}

type Manager struct {
	mu       sync.Mutex
	redis    redis.Cmdable
	adapters map[string]exchange.ExchangeAdapter
	logger   *slog.Logger
}

func NewManager(
	rds redis.Cmdable,
	adapters map[string]exchange.ExchangeAdapter,
	logger *slog.Logger,
) *Manager {
	return &Manager{
		redis:    rds,
		adapters: adapters,
		logger:   logger,
	}
}

func (m *Manager) Add(ctx context.Context, req SubscribeRequest) (SubscribeResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	targets := m.resolveTargets(req.Exchanges)
	result := SubscribeResult{Failed: make(map[string]string)}

	for _, name := range targets {
		key := fmt.Sprintf("sub:%s", name)
		members := make([]any, len(req.Symbols))
		for i, s := range req.Symbols {
			members[i] = s
		}
		if err := m.redis.SAdd(ctx, key, members...).Err(); err != nil {
			result.Failed[name] = err.Error()
			continue
		}

		adapter, ok := m.adapters[name]
		if !ok {
			result.Failed[name] = "adapter not found"
			continue
		}
		if err := adapter.Subscribe(req.Symbols); err != nil {
			result.Failed[name] = err.Error()
			continue
		}
		result.Succeeded = append(result.Succeeded, name)
	}

	return result, nil
}

func (m *Manager) Remove(ctx context.Context, req UnsubscribeRequest) (SubscribeResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	targets := m.resolveTargets(req.Exchanges)
	result := SubscribeResult{Failed: make(map[string]string)}

	for _, name := range targets {
		key := fmt.Sprintf("sub:%s", name)
		members := make([]any, len(req.Symbols))
		for i, s := range req.Symbols {
			members[i] = s
		}
		if err := m.redis.SRem(ctx, key, members...).Err(); err != nil {
			result.Failed[name] = err.Error()
			continue
		}

		adapter, ok := m.adapters[name]
		if !ok {
			result.Failed[name] = "adapter not found"
			continue
		}
		if err := adapter.Unsubscribe(req.Symbols); err != nil {
			result.Failed[name] = err.Error()
			continue
		}
		result.Succeeded = append(result.Succeeded, name)
	}

	return result, nil
}

func (m *Manager) List(ctx context.Context) ([]SubscriptionInfo, error) {
	var infos []SubscriptionInfo
	for name := range m.adapters {
		key := fmt.Sprintf("sub:%s", name)
		symbols, err := m.redis.SMembers(ctx, key).Result()
		if err != nil {
			return nil, fmt.Errorf("list %s: %w", name, err)
		}
		if len(symbols) > 0 {
			infos = append(infos, SubscriptionInfo{Exchange: name, Symbols: symbols})
		}
	}
	return infos, nil
}

func (m *Manager) Restore(ctx context.Context) error {
	for name, adapter := range m.adapters {
		key := fmt.Sprintf("sub:%s", name)
		symbols, err := m.redis.SMembers(ctx, key).Result()
		if err != nil {
			m.logger.Warn("restore subscriptions", "exchange", name, "error", err)
			continue
		}
		if len(symbols) > 0 {
			if err := adapter.Subscribe(symbols); err != nil {
				m.logger.Warn("restore subscribe", "exchange", name, "error", err)
			}
		}
	}
	return nil
}

func (m *Manager) resolveTargets(exchanges []string) []string {
	if len(exchanges) == 0 {
		targets := make([]string, 0, len(m.adapters))
		for name := range m.adapters {
			targets = append(targets, name)
		}
		return targets
	}
	return exchanges
}
