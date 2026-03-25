package exchange

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/yanking/price-watch/internal/watch/event"
	"github.com/yanking/price-watch/pkg/eventbus"
)

type ExchangeAdapter interface {
	Name() string
	Subscribe(symbols []string) error
	Unsubscribe(symbols []string) error
	FetchKlines(ctx context.Context, symbol, interval string, start, end time.Time) ([]event.KlineData, error)
	Start(ctx context.Context) error
	Stop() error
}

type MessageParser func(msg []byte) (any, error)
type MessageBuilder func(symbols []string) ([]byte, error)
type PingBuilder func() []byte

type BaseAdapter struct {
	name         string
	bus          eventbus.Bus
	logger       *slog.Logger
	parser       MessageParser
	subBuilder   MessageBuilder
	unsubBuilder MessageBuilder
	pingBuilder  PingBuilder
	cfg          ExchangeConfig

	mu      sync.Mutex
	conn    *websocket.Conn
	ctx     context.Context
	cancel  context.CancelFunc
	symbols []string
}

func NewBaseAdapter(
	name string,
	bus eventbus.Bus,
	logger *slog.Logger,
	cfg ExchangeConfig,
	parser MessageParser,
	subBuilder MessageBuilder,
	unsubBuilder MessageBuilder,
	pingBuilder PingBuilder,
) *BaseAdapter {
	return &BaseAdapter{
		name:         name,
		bus:          bus,
		logger:       logger.With("exchange", name),
		cfg:          cfg,
		parser:       parser,
		subBuilder:   subBuilder,
		unsubBuilder: unsubBuilder,
		pingBuilder:  pingBuilder,
	}
}

func (b *BaseAdapter) Name() string { return b.name }

func (b *BaseAdapter) Start(ctx context.Context) error {
	b.ctx, b.cancel = context.WithCancel(ctx)
	if err := b.connect(); err != nil {
		return err
	}
	go b.readLoop()
	go b.pingLoop()
	return nil
}

func (b *BaseAdapter) Stop() error {
	if b.cancel != nil {
		b.cancel()
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.conn != nil {
		return b.conn.Close()
	}
	return nil
}

func (b *BaseAdapter) Subscribe(symbols []string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.symbols = append(b.symbols, symbols...)
	if b.conn == nil {
		return nil
	}
	msg, err := b.subBuilder(symbols)
	if err != nil {
		return fmt.Errorf("build subscribe msg: %w", err)
	}
	return b.conn.WriteMessage(websocket.TextMessage, msg)
}

func (b *BaseAdapter) Unsubscribe(symbols []string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	msg, err := b.unsubBuilder(symbols)
	if err != nil {
		return fmt.Errorf("build unsubscribe msg: %w", err)
	}
	removeSet := make(map[string]bool, len(symbols))
	for _, s := range symbols {
		removeSet[s] = true
	}
	remaining := make([]string, 0, len(b.symbols))
	for _, s := range b.symbols {
		if !removeSet[s] {
			remaining = append(remaining, s)
		}
	}
	b.symbols = remaining
	if b.conn == nil {
		return nil
	}
	return b.conn.WriteMessage(websocket.TextMessage, msg)
}

func (b *BaseAdapter) connect() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	conn, _, err := websocket.DefaultDialer.Dial(b.cfg.WsURL, nil)
	if err != nil {
		return fmt.Errorf("connect %s: %w", b.name, err)
	}
	b.conn = conn
	if len(b.symbols) > 0 {
		msg, err := b.subBuilder(b.symbols)
		if err != nil {
			return fmt.Errorf("build resubscribe msg: %w", err)
		}
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return fmt.Errorf("resubscribe: %w", err)
		}
	}
	return nil
}

func (b *BaseAdapter) reconnect() {
	attempt := 0
	for {
		select {
		case <-b.ctx.Done():
			return
		default:
		}
		delay := time.Duration(math.Min(
			float64(b.cfg.ReconnectBase)*math.Pow(2, float64(attempt)),
			float64(b.cfg.ReconnectMax),
		))
		b.logger.Info("reconnecting", "attempt", attempt+1, "delay", delay)
		time.Sleep(delay)
		if err := b.connect(); err != nil {
			b.logger.Warn("reconnect failed", "error", err)
			attempt++
			continue
		}
		b.logger.Info("reconnected")
		return
	}
}

func (b *BaseAdapter) readLoop() {
	for {
		select {
		case <-b.ctx.Done():
			return
		default:
		}
		b.mu.Lock()
		conn := b.conn
		b.mu.Unlock()
		if conn == nil {
			b.reconnect()
			continue
		}
		_, msg, err := conn.ReadMessage()
		if err != nil {
			b.logger.Warn("read error", "error", err)
			b.mu.Lock()
			b.conn = nil
			b.mu.Unlock()
			b.reconnect()
			continue
		}
		parsed, err := b.parser(msg)
		if err != nil {
			continue
		}
		b.publishParsed(parsed)
	}
}

func (b *BaseAdapter) publishParsed(parsed any) {
	switch v := parsed.(type) {
	case event.Event[event.TickData]:
		data, err := event.Marshal(v)
		if err != nil {
			b.logger.Warn("marshal tick event", "error", err)
			return
		}
		_ = b.bus.Publish(b.ctx, v.BuildSubject(), data)
	case event.Event[event.KlineData]:
		data, err := event.Marshal(v)
		if err != nil {
			b.logger.Warn("marshal kline event", "error", err)
			return
		}
		_ = b.bus.Publish(b.ctx, v.BuildSubject(), data)
	}
}

func (b *BaseAdapter) pingLoop() {
	ticker := time.NewTicker(b.cfg.PingInterval)
	defer ticker.Stop()
	for {
		select {
		case <-b.ctx.Done():
			return
		case <-ticker.C:
			b.mu.Lock()
			conn := b.conn
			b.mu.Unlock()
			if conn == nil {
				continue
			}
			payload := b.pingBuilder()
			if payload != nil {
				_ = conn.WriteMessage(websocket.TextMessage, payload)
			} else {
				_ = conn.WriteMessage(websocket.PingMessage, nil)
			}
		}
	}
}

// NewAdapter creates the appropriate exchange adapter by name.
// Returns nil for unknown exchange names.
func NewAdapter(name string, bus eventbus.Bus, logger *slog.Logger, cfg ExchangeConfig) ExchangeAdapter {
	switch name {
	case "binance":
		return NewBinanceAdapter(bus, logger, cfg)
	case "okx":
		return NewOKXAdapter(bus, logger, cfg)
	case "bybit":
		return NewBybitAdapter(bus, logger, cfg)
	default:
		return nil
	}
}
