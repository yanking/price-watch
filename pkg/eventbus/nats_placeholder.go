package eventbus

import (
	"context"
	"fmt"
	"log/slog"
)

type natsBus struct{}

func newNATSBus(cfg NATSConfig, logger *slog.Logger) (*natsBus, error) {
	return nil, fmt.Errorf("nats eventbus not implemented yet")
}

func (b *natsBus) Publish(ctx context.Context, subject string, data []byte) error { return nil }
func (b *natsBus) Subscribe(subject string, handler Handler) (Subscription, error) { return nil, nil }
func (b *natsBus) Drain() error { return nil }
func (b *natsBus) Close() error { return nil }
