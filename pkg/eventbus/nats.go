package eventbus

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type natsBus struct {
	conn   *nats.Conn
	js     jetstream.JetStream
	logger *slog.Logger
	subs   []jetstream.ConsumeContext
}

func newNATSBus(cfg NATSConfig, logger *slog.Logger) (*natsBus, error) {
	nc, err := nats.Connect(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("connect nats: %w", err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("create jetstream: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, sc := range cfg.Streams {
		_, err := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
			Name:     sc.Name,
			Subjects: sc.Subjects,
			MaxAge:   sc.MaxAge.Std(),
		})
		if err != nil {
			nc.Close()
			return nil, fmt.Errorf("create stream %s: %w", sc.Name, err)
		}
	}

	return &natsBus{conn: nc, js: js, logger: logger}, nil
}

func (b *natsBus) Publish(ctx context.Context, subject string, data []byte) error {
	_, err := b.js.Publish(ctx, subject, data)
	if err != nil {
		return fmt.Errorf("publish to %s: %w", subject, err)
	}
	return nil
}

func (b *natsBus) Subscribe(subject string, handler Handler) (Subscription, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	streamName := ""
	for _, stream := range b.listStreams() {
		info, err := b.js.Stream(ctx, stream)
		if err != nil {
			continue
		}
		cfg := info.CachedInfo().Config
		for _, s := range cfg.Subjects {
			if matchSubject(s, subject) || s == subject {
				streamName = stream
				break
			}
		}
		if streamName != "" {
			break
		}
	}

	if streamName == "" {
		return nil, fmt.Errorf("no stream found for subject %q", subject)
	}

	cons, err := b.js.CreateOrUpdateConsumer(ctx, streamName, jetstream.ConsumerConfig{
		FilterSubject: subject,
		DeliverPolicy: jetstream.DeliverNewPolicy,
		AckPolicy:     jetstream.AckExplicitPolicy,
	})
	if err != nil {
		return nil, fmt.Errorf("create consumer: %w", err)
	}

	cc, err := cons.Consume(func(msg jetstream.Msg) {
		if hErr := handler(msg.Subject(), msg.Data()); hErr != nil {
			b.logger.Warn("handler error", "subject", msg.Subject(), "error", hErr)
		}
		_ = msg.Ack()
	})
	if err != nil {
		return nil, fmt.Errorf("start consume: %w", err)
	}

	b.subs = append(b.subs, cc)
	return &natsSubscription{cc: cc}, nil
}

func (b *natsBus) Drain() error {
	for _, s := range b.subs {
		s.Stop()
	}
	return b.conn.Drain()
}

func (b *natsBus) Close() error {
	for _, s := range b.subs {
		s.Stop()
	}
	b.conn.Close()
	return nil
}

func (b *natsBus) listStreams() []string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var names []string
	sl := b.js.ListStreams(ctx)
	for info := range sl.Info() {
		names = append(names, info.Config.Name)
	}
	return names
}

type natsSubscription struct {
	cc jetstream.ConsumeContext
}

func (s *natsSubscription) Unsubscribe() error {
	s.cc.Stop()
	return nil
}
