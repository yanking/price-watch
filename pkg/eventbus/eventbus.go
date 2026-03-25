package eventbus

import (
	"context"
	"fmt"
	"log/slog"
)

type Bus interface {
	Publish(ctx context.Context, subject string, data []byte) error
	Subscribe(subject string, handler Handler) (Subscription, error)
	Drain() error
	Close() error
}

type Handler func(subject string, data []byte) error

type Subscription interface {
	Unsubscribe() error
}

func New(cfg Config, logger *slog.Logger) (Bus, error) {
	switch cfg.Driver {
	case "memory":
		return newMemoryBus(cfg.Memory)
	case "nats":
		return newNATSBus(cfg.NATS, logger)
	default:
		return nil, fmt.Errorf("unsupported eventbus driver: %q", cfg.Driver)
	}
}
