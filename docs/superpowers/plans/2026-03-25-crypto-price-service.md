# Crypto Price Service Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a real-time cryptocurrency price fetching service that connects to 4 exchanges via WebSocket, stores data in Redis (cache) and InfluxDB (history), and exposes REST APIs for querying.

**Architecture:** Event-driven with pluggable EventBus (in-memory or NATS JetStream). Exchange adapters publish unified events; independent subscribers consume them for caching and storage. HTTP API provides price queries, K-line queries, and dynamic subscription management.

**Tech Stack:** Go 1.26.1, Gin (HTTP), gorilla/websocket (WebSocket), go-redis v9, InfluxDB client v2, NATS JetStream, shopspring/decimal, oklog/ulid

**Spec:** `docs/superpowers/specs/2026-03-25-crypto-price-service-design.md`

---

## File Map

### New packages (pkg/)

| File | Responsibility |
|------|---------------|
| `pkg/eventbus/eventbus.go` | `Bus` interface, `Handler`, `Subscription`, `New()` factory |
| `pkg/eventbus/config.go` | `Config`, `MemoryConfig`, `NATSConfig`, `StreamConfig` structs |
| `pkg/eventbus/memory.go` | In-memory Bus implementation with wildcard matching |
| `pkg/eventbus/memory_test.go` | Tests for in-memory Bus |
| `pkg/eventbus/nats.go` | NATS JetStream Bus implementation |
| `pkg/eventbus/nats_test.go` | Tests for NATS Bus (using embedded NATS server) |
| `pkg/database/influxdb/config.go` | `Config` struct with Validate |
| `pkg/database/influxdb/influxdb.go` | `Client` wrapper (write, query, close) |
| `pkg/database/influxdb/influxdb_test.go` | Tests for InfluxDB client |

### New packages (internal/watch/)

| File | Responsibility |
|------|---------------|
| `internal/watch/event/event.go` | `Event[T]`, `TickData`, `KlineData`, `BuildSubject()` |
| `internal/watch/event/marshal.go` | `Marshal`, `Unmarshal`, `UnmarshalData` helpers |
| `internal/watch/event/event_test.go` | Tests for event model and serialization |
| `internal/watch/event/publisher.go` | `Publisher` helper — wraps Bus, builds subject, marshals events |
| `internal/watch/event/subscriber.go` | Typed subscribe helpers — `SubscribeTick`, `SubscribeKline` |
| `internal/watch/exchange/adapter.go` | `ExchangeAdapter` interface, `BaseAdapter`, callback types |
| `internal/watch/exchange/config.go` | `ExchangeConfig` struct |
| `internal/watch/exchange/binance.go` | Binance adapter |
| `internal/watch/exchange/binance_test.go` | Binance parser/builder tests |
| `internal/watch/exchange/okx.go` | OKX adapter |
| `internal/watch/exchange/okx_test.go` | OKX parser/builder tests |
| `internal/watch/exchange/bybit.go` | Bybit adapter |
| `internal/watch/exchange/bybit_test.go` | Bybit parser/builder tests |
| `internal/watch/exchange/gateio.go` | Gate.io adapter |
| `internal/watch/exchange/gateio_test.go` | Gate.io parser/builder tests |
| `internal/watch/subscription/manager.go` | `SubscriptionManager` implementation |
| `internal/watch/subscription/manager_test.go` | Tests with miniredis |
| `internal/watch/consumer/cache.go` | `CacheSubscriber` → Redis |
| `internal/watch/consumer/cache_test.go` | Tests with miniredis |
| `internal/watch/consumer/storage.go` | `StorageSubscriber` → InfluxDB |
| `internal/watch/consumer/storage_test.go` | Tests with mock InfluxDB writer |
| `internal/watch/handler/response.go` | Unified response struct and helpers |
| `internal/watch/handler/subscription.go` | Subscription CRUD handlers |
| `internal/watch/handler/price.go` | Price query handlers |
| `internal/watch/handler/kline.go` | K-line query handler |
| `internal/watch/handler/health.go` | Health/ready endpoints |
| `internal/watch/handler/routes.go` | `RegisterRoutes` wires all handlers to Gin engine |
| `internal/watch/handler/handler_test.go` | HTTP handler tests with httptest |
| `internal/watch/server/http.go` | `GinServer` implementing `app.Server` |

### Modified files

| File | Changes |
|------|---------|
| `internal/watch/config/config.go` | Remove MySQL, add InfluxDB/EventBus/HTTP/Exchanges sections |
| `internal/watch/config/config_test.go` | Update to match new config struct |
| `internal/watch/svc/serviceContext.go` | Remove MySQL, add Influx/Bus/SubMgr/Adapters |
| `internal/watch/svc/serviceContext_test.go` | Update to match new ServiceContext |
| `cmd/watch/initial/initApp.go` | Start consumers, restore subscriptions |
| `cmd/watch/initial/createService.go` | Register GinServer + adapter wrappers |
| `cmd/watch/initial/close.go` | Ordered cleanup registration |
| `configs/watch.yaml` | New config structure per spec |
| `docker-compose.yml` | Add InfluxDB + NATS services |
| `.env.example` | Add InfluxDB + NATS env vars |
| `go.mod` | New dependencies |

---

## Task 1: Add new dependencies

**Files:**
- Modify: `go.mod`

- [ ] **Step 1: Add all required dependencies**

```bash
cd D:/data/ai/price-watch
go get github.com/gin-gonic/gin@latest
go get github.com/gorilla/websocket@latest
go get github.com/shopspring/decimal@latest
go get github.com/oklog/ulid/v2@latest
go get github.com/influxdata/influxdb-client-go/v2@latest
go get github.com/nats-io/nats.go@latest
go get github.com/nats-io/nats-server/v2@latest
go get golang.org/x/time@latest
```

- [ ] **Step 2: Tidy modules**

```bash
go mod tidy
```

- [ ] **Step 3: Verify compilation**

```bash
go build ./...
```

Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: add dependencies for crypto price service"
```

---

## Task 2: EventBus — interface and in-memory implementation

**Files:**
- Create: `pkg/eventbus/eventbus.go`
- Create: `pkg/eventbus/config.go`
- Create: `pkg/eventbus/memory.go`
- Create: `pkg/eventbus/memory_test.go`

- [ ] **Step 1: Write failing tests for in-memory Bus**

Create `pkg/eventbus/memory_test.go`:

```go
package eventbus

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestMemoryBus_PublishSubscribe(t *testing.T) {
	bus, err := newMemoryBus(MemoryConfig{BufferSize: 64})
	if err != nil {
		t.Fatalf("newMemoryBus: %v", err)
	}
	defer bus.Close()

	var received []byte
	var mu sync.Mutex
	done := make(chan struct{})

	_, err = bus.Subscribe("test.topic", func(subject string, data []byte) error {
		mu.Lock()
		received = append([]byte{}, data...)
		mu.Unlock()
		close(done)
		return nil
	})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	err = bus.Publish(context.Background(), "test.topic", []byte("hello"))
	if err != nil {
		t.Fatalf("Publish: %v", err)
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for message")
	}

	mu.Lock()
	defer mu.Unlock()
	if string(received) != "hello" {
		t.Errorf("got %q, want %q", received, "hello")
	}
}

func TestMemoryBus_WildcardStar(t *testing.T) {
	bus, err := newMemoryBus(MemoryConfig{BufferSize: 64})
	if err != nil {
		t.Fatalf("newMemoryBus: %v", err)
	}
	defer bus.Close()

	var count int
	var mu sync.Mutex
	done := make(chan struct{}, 2)

	_, err = bus.Subscribe("price.tick.*", func(subject string, data []byte) error {
		mu.Lock()
		count++
		mu.Unlock()
		done <- struct{}{}
		return nil
	})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	bus.Publish(context.Background(), "price.tick.binance", []byte("1"))
	bus.Publish(context.Background(), "price.tick.okx", []byte("2"))
	bus.Publish(context.Background(), "price.kline.binance", []byte("3")) // should NOT match

	for i := 0; i < 2; i++ {
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("timed out")
		}
	}

	// Give time for non-matching message to (not) arrive
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if count != 2 {
		t.Errorf("got %d messages, want 2", count)
	}
}

func TestMemoryBus_WildcardGreaterThan(t *testing.T) {
	bus, err := newMemoryBus(MemoryConfig{BufferSize: 64})
	if err != nil {
		t.Fatalf("newMemoryBus: %v", err)
	}
	defer bus.Close()

	var count int
	var mu sync.Mutex
	done := make(chan struct{}, 3)

	_, err = bus.Subscribe("price.tick.>", func(subject string, data []byte) error {
		mu.Lock()
		count++
		mu.Unlock()
		done <- struct{}{}
		return nil
	})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	bus.Publish(context.Background(), "price.tick.binance.BTCUSDT", []byte("1"))
	bus.Publish(context.Background(), "price.tick.okx.ETHUSDT", []byte("2"))
	bus.Publish(context.Background(), "price.kline.binance.BTCUSDT", []byte("3")) // should NOT match

	for i := 0; i < 2; i++ {
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("timed out")
		}
	}

	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if count != 2 {
		t.Errorf("got %d messages, want 2", count)
	}
}

func TestMemoryBus_Unsubscribe(t *testing.T) {
	bus, err := newMemoryBus(MemoryConfig{BufferSize: 64})
	if err != nil {
		t.Fatalf("newMemoryBus: %v", err)
	}
	defer bus.Close()

	var count int
	var mu sync.Mutex

	sub, err := bus.Subscribe("test", func(subject string, data []byte) error {
		mu.Lock()
		count++
		mu.Unlock()
		return nil
	})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	bus.Publish(context.Background(), "test", []byte("1"))
	time.Sleep(50 * time.Millisecond)

	sub.Unsubscribe()

	bus.Publish(context.Background(), "test", []byte("2"))
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if count != 1 {
		t.Errorf("got %d messages, want 1 (after unsubscribe)", count)
	}
}

func TestMemoryBus_Drain(t *testing.T) {
	bus, err := newMemoryBus(MemoryConfig{BufferSize: 64})
	if err != nil {
		t.Fatalf("newMemoryBus: %v", err)
	}

	var count int
	var mu sync.Mutex

	_, err = bus.Subscribe("test", func(subject string, data []byte) error {
		mu.Lock()
		count++
		mu.Unlock()
		return nil
	})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	bus.Publish(context.Background(), "test", []byte("1"))
	bus.Publish(context.Background(), "test", []byte("2"))

	err = bus.Drain()
	if err != nil {
		t.Fatalf("Drain: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if count != 2 {
		t.Errorf("got %d messages after drain, want 2", count)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./pkg/eventbus/... -v
```

Expected: compilation errors (types not defined yet)

- [ ] **Step 3: Write Bus interface and config**

Create `pkg/eventbus/eventbus.go`:

```go
package eventbus

import (
	"context"
	"fmt"
	"log/slog"
)

// Bus is a generic publish/subscribe message bus.
// Implementations must be safe for concurrent use.
type Bus interface {
	Publish(ctx context.Context, subject string, data []byte) error
	Subscribe(subject string, handler Handler) (Subscription, error)
	Drain() error
	Close() error
}

// Handler processes a message delivered to a subscription.
type Handler func(subject string, data []byte) error

// Subscription represents an active subscription that can be cancelled.
type Subscription interface {
	Unsubscribe() error
}

// New creates a Bus from config.
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
```

Create `pkg/eventbus/config.go`:

```go
package eventbus

import (
	"fmt"
	"time"
)

type Config struct {
	Driver string       `mapstructure:"driver"`
	Memory MemoryConfig `mapstructure:"memory"`
	NATS   NATSConfig   `mapstructure:"nats"`
}

func (c *Config) Validate() error {
	switch c.Driver {
	case "memory":
		return c.Memory.Validate()
	case "nats":
		return c.NATS.Validate()
	default:
		return fmt.Errorf("unsupported eventbus driver: %q", c.Driver)
	}
}

type MemoryConfig struct {
	BufferSize int `mapstructure:"bufferSize"`
}

func (c *MemoryConfig) Validate() error {
	if c.BufferSize <= 0 {
		c.BufferSize = 1024
	}
	return nil
}

type NATSConfig struct {
	URL     string         `mapstructure:"url"`
	Streams []StreamConfig `mapstructure:"streams"`
}

func (c *NATSConfig) Validate() error {
	if c.URL == "" {
		return fmt.Errorf("nats url is required")
	}
	return nil
}

type StreamConfig struct {
	Name     string   `mapstructure:"name"`
	Subjects []string `mapstructure:"subjects"`
	MaxAge   Duration `mapstructure:"maxAge"`
}

// Duration wraps time.Duration for mapstructure string decoding.
type Duration time.Duration

func (d *Duration) UnmarshalText(text []byte) error {
	dur, err := time.ParseDuration(string(text))
	if err != nil {
		return err
	}
	*d = Duration(dur)
	return nil
}

func (d Duration) Std() time.Duration {
	return time.Duration(d)
}
```

- [ ] **Step 4: Write in-memory Bus implementation**

Create `pkg/eventbus/memory.go`:

```go
package eventbus

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

type memoryBus struct {
	mu         sync.RWMutex
	subs       []*memorySub
	bufferSize int
	closed     bool
}

type memorySub struct {
	pattern string
	ch      chan message
	handler Handler
	done    chan struct{}
	once    sync.Once
}

type message struct {
	subject string
	data    []byte
}

func newMemoryBus(cfg MemoryConfig) (*memoryBus, error) {
	_ = cfg.Validate()
	return &memoryBus{
		bufferSize: cfg.BufferSize,
	}, nil
}

func (b *memoryBus) Publish(ctx context.Context, subject string, data []byte) error {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.closed {
		return fmt.Errorf("bus is closed")
	}
	msg := message{subject: subject, data: data}
	for _, s := range b.subs {
		if matchSubject(s.pattern, subject) {
			select {
			case s.ch <- msg:
			default:
				// slow consumer, drop message
			}
		}
	}
	return nil
}

func (b *memoryBus) Subscribe(pattern string, handler Handler) (Subscription, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return nil, fmt.Errorf("bus is closed")
	}
	s := &memorySub{
		pattern: pattern,
		ch:      make(chan message, b.bufferSize),
		handler: handler,
		done:    make(chan struct{}),
	}
	b.subs = append(b.subs, s)
	go s.loop()
	return &memorySubscription{bus: b, sub: s}, nil
}

func (b *memoryBus) Drain() error {
	b.mu.Lock()
	b.closed = true
	subs := make([]*memorySub, len(b.subs))
	copy(subs, b.subs)
	b.mu.Unlock()

	for _, s := range subs {
		close(s.ch)
		<-s.done
	}
	return nil
}

func (b *memoryBus) Close() error {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return nil
	}
	b.closed = true
	subs := make([]*memorySub, len(b.subs))
	copy(subs, b.subs)
	b.subs = nil
	b.mu.Unlock()

	for _, s := range subs {
		s.once.Do(func() { close(s.ch) })
		<-s.done
	}
	return nil
}

func (b *memoryBus) removeSub(sub *memorySub) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for i, s := range b.subs {
		if s == sub {
			b.subs = append(b.subs[:i], b.subs[i+1:]...)
			break
		}
	}
}

func (s *memorySub) loop() {
	defer close(s.done)
	for msg := range s.ch {
		_ = s.handler(msg.subject, msg.data)
	}
}

type memorySubscription struct {
	bus *memoryBus
	sub *memorySub
}

func (ms *memorySubscription) Unsubscribe() error {
	ms.bus.removeSub(ms.sub)
	ms.sub.once.Do(func() { close(ms.sub.ch) })
	<-ms.sub.done
	return nil
}

// matchSubject implements NATS-style wildcard matching.
// '*' matches exactly one token; '>' matches one or more tokens (must be last).
func matchSubject(pattern, subject string) bool {
	patParts := strings.Split(pattern, ".")
	subParts := strings.Split(subject, ".")

	for i, p := range patParts {
		if p == ">" {
			return i < len(subParts)
		}
		if i >= len(subParts) {
			return false
		}
		if p != "*" && p != subParts[i] {
			return false
		}
	}
	return len(patParts) == len(subParts)
}
```

(The `"fmt"` import is already included in the code above.)

- [ ] **Step 5: Run tests to verify they pass**

```bash
go test ./pkg/eventbus/... -v
```

Expected: all PASS

- [ ] **Step 6: Commit**

```bash
git add pkg/eventbus/
git commit -m "feat: add eventbus package with in-memory implementation"
```

---

## Task 3: EventBus — NATS JetStream implementation

**Files:**
- Create: `pkg/eventbus/nats.go`
- Create: `pkg/eventbus/nats_test.go`

- [ ] **Step 1: Write failing tests for NATS Bus**

Create `pkg/eventbus/nats_test.go`:

```go
package eventbus

import (
	"context"
	"log/slog"
	"sync"
	"testing"
	"time"

	natsserver "github.com/nats-io/nats-server/v2/server"
)

func startTestNATS(t *testing.T) (*natsserver.Server, string) {
	t.Helper()
	opts := &natsserver.Options{
		Port:      -1,
		JetStream: true,
		StoreDir:  t.TempDir(),
	}
	ns, err := natsserver.NewServer(opts)
	if err != nil {
		t.Fatalf("start nats server: %v", err)
	}
	ns.Start()
	if !ns.ReadyForConnections(5 * time.Second) {
		t.Fatal("nats server not ready")
	}
	return ns, ns.ClientURL()
}

func TestNATSBus_PublishSubscribe(t *testing.T) {
	ns, url := startTestNATS(t)
	defer ns.Shutdown()

	cfg := NATSConfig{
		URL: url,
		Streams: []StreamConfig{
			{Name: "TEST", Subjects: []string{"test.>"}, MaxAge: Duration(time.Hour)},
		},
	}

	bus, err := newNATSBus(cfg, slog.Default())
	if err != nil {
		t.Fatalf("newNATSBus: %v", err)
	}
	defer bus.Close()

	var received []byte
	var mu sync.Mutex
	done := make(chan struct{})

	_, err = bus.Subscribe("test.>", func(subject string, data []byte) error {
		mu.Lock()
		received = append([]byte{}, data...)
		mu.Unlock()
		close(done)
		return nil
	})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	// Small delay for subscription to be ready
	time.Sleep(200 * time.Millisecond)

	err = bus.Publish(context.Background(), "test.hello", []byte("world"))
	if err != nil {
		t.Fatalf("Publish: %v", err)
	}

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for message")
	}

	mu.Lock()
	defer mu.Unlock()
	if string(received) != "world" {
		t.Errorf("got %q, want %q", received, "world")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./pkg/eventbus/... -run TestNATS -v
```

Expected: compilation error (newNATSBus not defined)

- [ ] **Step 3: Write NATS Bus implementation**

Create `pkg/eventbus/nats.go`:

```go
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

	// Find which stream covers this subject
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
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./pkg/eventbus/... -v
```

Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add pkg/eventbus/nats.go pkg/eventbus/nats_test.go
git commit -m "feat: add NATS JetStream eventbus implementation"
```

---

## Task 4: InfluxDB client package

**Files:**
- Create: `pkg/database/influxdb/config.go`
- Create: `pkg/database/influxdb/influxdb.go`
- Create: `pkg/database/influxdb/influxdb_test.go`

Follow the same pattern as `pkg/database/redisx/` (Config struct + Validate + Client wrapper).

- [ ] **Step 1: Write failing test**

Create `pkg/database/influxdb/influxdb_test.go`:

```go
package influxdb

import (
	"testing"
	"time"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name:    "empty URL",
			cfg:     Config{},
			wantErr: true,
		},
		{
			name: "empty org",
			cfg: Config{
				URL: "http://localhost:8086",
			},
			wantErr: true,
		},
		{
			name: "valid config",
			cfg: Config{
				URL:           "http://localhost:8086",
				Token:         "test-token",
				Org:           "test-org",
				Buckets:       BucketsConfig{Tick: "prices-tick", Kline: "prices-kline"},
				BatchSize:     500,
				FlushInterval: 5 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "defaults applied",
			cfg: Config{
				URL:     "http://localhost:8086",
				Org:     "test-org",
				Buckets: BucketsConfig{Tick: "prices-tick", Kline: "prices-kline"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./pkg/database/influxdb/... -v
```

Expected: compilation error

- [ ] **Step 3: Write config**

Create `pkg/database/influxdb/config.go`:

```go
package influxdb

import (
	"fmt"
	"time"
)

type Config struct {
	URL           string        `mapstructure:"url"`
	Token         string        `mapstructure:"token"`
	Org           string        `mapstructure:"org"`
	Buckets       BucketsConfig `mapstructure:"buckets"`
	BatchSize     uint          `mapstructure:"batchSize"`
	FlushInterval time.Duration `mapstructure:"flushInterval"`
}

type BucketsConfig struct {
	Tick  string `mapstructure:"tick"`
	Kline string `mapstructure:"kline"`
}

func (c *Config) Validate() error {
	if c.URL == "" {
		return fmt.Errorf("url is required")
	}
	if c.Org == "" {
		return fmt.Errorf("org is required")
	}
	if c.Buckets.Tick == "" {
		return fmt.Errorf("buckets.tick is required")
	}
	if c.Buckets.Kline == "" {
		return fmt.Errorf("buckets.kline is required")
	}
	if c.BatchSize == 0 {
		c.BatchSize = 500
	}
	if c.FlushInterval == 0 {
		c.FlushInterval = 5 * time.Second
	}
	return nil
}
```

- [ ] **Step 4: Write client**

Create `pkg/database/influxdb/influxdb.go`:

```go
package influxdb

import (
	"context"
	"fmt"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

type Client struct {
	client      influxdb2.Client
	tickWriter  api.WriteAPIBlocking
	klineWriter api.WriteAPIBlocking
	queryAPI    api.QueryAPI
	cfg         Config
}

func New(cfg Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	client := influxdb2.NewClient(cfg.URL, cfg.Token)

	return &Client{
		client:      client,
		tickWriter:  client.WriteAPIBlocking(cfg.Org, cfg.Buckets.Tick),
		klineWriter: client.WriteAPIBlocking(cfg.Org, cfg.Buckets.Kline),
		queryAPI:    client.QueryAPI(cfg.Org),
		cfg:         cfg,
	}, nil
}

func (c *Client) WriteTickPoints(ctx context.Context, points ...*write.Point) error {
	return c.tickWriter.WritePoint(ctx, points...)
}

func (c *Client) WriteKlinePoints(ctx context.Context, points ...*write.Point) error {
	return c.klineWriter.WritePoint(ctx, points...)
}

func (c *Client) Query(ctx context.Context, query string) (*api.QueryTableResult, error) {
	return c.queryAPI.Query(ctx, query)
}

func (c *Client) Ping(ctx context.Context) error {
	ok, err := c.client.Ping(ctx)
	if err != nil {
		return fmt.Errorf("ping influxdb: %w", err)
	}
	if !ok {
		return fmt.Errorf("influxdb ping returned false")
	}
	return nil
}

func (c *Client) Close() error {
	c.client.Close()
	return nil
}

func (c *Client) Config() Config {
	return c.cfg
}
```

- [ ] **Step 5: Run tests**

```bash
go test ./pkg/database/influxdb/... -v
```

Expected: config tests PASS

- [ ] **Step 6: Commit**

```bash
git add pkg/database/influxdb/
git commit -m "feat: add influxdb client package"
```

---

## Task 5: Event model and serialization

**Files:**
- Create: `internal/watch/event/event.go`
- Create: `internal/watch/event/marshal.go`
- Create: `internal/watch/event/event_test.go`

- [ ] **Step 1: Write failing tests**

Create `internal/watch/event/event_test.go`:

```go
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
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/watch/event/... -v
```

Expected: compilation error

- [ ] **Step 3: Write event model**

Create `internal/watch/event/event.go`:

```go
package event

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/shopspring/decimal"
)

const (
	TypeTick  = "price.tick"
	TypeKline = "price.kline"
)

type Event[T any] struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Source    string    `json:"source"`
	Subject   string    `json:"subject"`
	Timestamp time.Time `json:"timestamp"`
	Data      T         `json:"data"`
}

func (e Event[T]) BuildSubject() string {
	return fmt.Sprintf("%s.%s.%s", e.Type, e.Source, e.Subject)
}

type TickData struct {
	Price  decimal.Decimal `json:"price"`
	Volume decimal.Decimal `json:"volume"`
}

type KlineData struct {
	Interval string          `json:"interval"`
	Open     decimal.Decimal `json:"open"`
	High     decimal.Decimal `json:"high"`
	Low      decimal.Decimal `json:"low"`
	Close    decimal.Decimal `json:"close"`
	Volume   decimal.Decimal `json:"volume"`
}

func NewTickEvent(source, symbol string, data TickData) Event[TickData] {
	return Event[TickData]{
		ID:        ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String(),
		Type:      TypeTick,
		Source:    source,
		Subject:   symbol,
		Timestamp: time.Now(),
		Data:      data,
	}
}

func NewKlineEvent(source, symbol string, data KlineData) Event[KlineData] {
	return Event[KlineData]{
		ID:        ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String(),
		Type:      TypeKline,
		Source:    source,
		Subject:   symbol,
		Timestamp: time.Now(),
		Data:      data,
	}
}
```

- [ ] **Step 4: Write serialization helpers**

Create `internal/watch/event/marshal.go`:

```go
package event

import (
	"encoding/json"
	"fmt"
)

func Marshal[T any](e Event[T]) ([]byte, error) {
	return json.Marshal(e)
}

func Unmarshal(data []byte) (eventType string, raw Event[json.RawMessage], err error) {
	if err = json.Unmarshal(data, &raw); err != nil {
		return "", raw, fmt.Errorf("unmarshal event envelope: %w", err)
	}
	return raw.Type, raw, nil
}

func UnmarshalData[T any](raw json.RawMessage) (T, error) {
	var v T
	if err := json.Unmarshal(raw, &v); err != nil {
		return v, fmt.Errorf("unmarshal event data: %w", err)
	}
	return v, nil
}
```

- [ ] **Step 5: Write publisher and subscriber helpers**

Create `internal/watch/event/publisher.go`:

```go
package event

import (
	"context"

	"github.com/yanking/price-watch/pkg/eventbus"
)

type Publisher struct {
	bus eventbus.Bus
}

func NewPublisher(bus eventbus.Bus) *Publisher {
	return &Publisher{bus: bus}
}

func (p *Publisher) PublishTick(ctx context.Context, e Event[TickData]) error {
	data, err := Marshal(e)
	if err != nil {
		return err
	}
	return p.bus.Publish(ctx, e.BuildSubject(), data)
}

func (p *Publisher) PublishKline(ctx context.Context, e Event[KlineData]) error {
	data, err := Marshal(e)
	if err != nil {
		return err
	}
	return p.bus.Publish(ctx, e.BuildSubject(), data)
}
```

Create `internal/watch/event/subscriber.go`:

```go
package event

import (
	"encoding/json"

	"github.com/yanking/price-watch/pkg/eventbus"
)

type TickHandler func(Event[TickData]) error

func SubscribeTick(bus eventbus.Bus, pattern string, h TickHandler) (eventbus.Subscription, error) {
	return bus.Subscribe(pattern, func(subject string, data []byte) error {
		_, raw, err := Unmarshal(data)
		if err != nil {
			return err
		}
		tick, err := UnmarshalData[TickData](raw.Data)
		if err != nil {
			return err
		}
		return h(Event[TickData]{
			ID: raw.ID, Type: raw.Type, Source: raw.Source,
			Subject: raw.Subject, Timestamp: raw.Timestamp, Data: tick,
		})
	})
}

type KlineHandler func(Event[KlineData]) error

func SubscribeKline(bus eventbus.Bus, pattern string, h KlineHandler) (eventbus.Subscription, error) {
	return bus.Subscribe(pattern, func(subject string, data []byte) error {
		_, raw, err := Unmarshal(data)
		if err != nil {
			return err
		}
		kline, err := UnmarshalData[KlineData](json.RawMessage(raw.Data))
		if err != nil {
			return err
		}
		return h(Event[KlineData]{
			ID: raw.ID, Type: raw.Type, Source: raw.Source,
			Subject: raw.Subject, Timestamp: raw.Timestamp, Data: kline,
		})
	})
}
```

- [ ] **Step 6: Run tests**

```bash
go test ./internal/watch/event/... -v
```

Expected: all PASS

- [ ] **Step 7: Commit**

```bash
git add internal/watch/event/
git commit -m "feat: add event model with CloudEvents-inspired envelope and serialization"
```

---

## Task 6: Update config and ServiceContext

**Files:**
- Modify: `internal/watch/config/config.go`
- Modify: `internal/watch/config/config_test.go`
- Modify: `internal/watch/svc/serviceContext.go`
- Modify: `internal/watch/svc/serviceContext_test.go`
- Modify: `configs/watch.yaml`
- Create: `internal/watch/exchange/config.go`

- [ ] **Step 1: Create exchange config**

Create `internal/watch/exchange/config.go`:

```go
package exchange

import (
	"fmt"
	"time"
)

type ExchangeConfig struct {
	Enabled       bool          `mapstructure:"enabled"`
	WsURL         string        `mapstructure:"wsUrl"`
	RestURL       string        `mapstructure:"restUrl"`
	ReconnectBase time.Duration `mapstructure:"reconnectBase"`
	ReconnectMax  time.Duration `mapstructure:"reconnectMax"`
	PingInterval  time.Duration `mapstructure:"pingInterval"`
}

func (c *ExchangeConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	if c.WsURL == "" {
		return fmt.Errorf("wsUrl is required")
	}
	if c.RestURL == "" {
		return fmt.Errorf("restUrl is required")
	}
	if c.ReconnectBase == 0 {
		c.ReconnectBase = time.Second
	}
	if c.ReconnectMax == 0 {
		c.ReconnectMax = 60 * time.Second
	}
	if c.PingInterval == 0 {
		c.PingInterval = 30 * time.Second
	}
	return nil
}

type HTTPConfig struct {
	Addr string `mapstructure:"addr"`
}

func (c *HTTPConfig) Validate() error {
	if c.Addr == "" {
		c.Addr = ":8080"
	}
	return nil
}
```

- [ ] **Step 2: Update config struct**

Replace `internal/watch/config/config.go` content:

```go
package config

import (
	"github.com/yanking/price-watch/pkg/database/influxdb"
	"github.com/yanking/price-watch/pkg/database/redisx"
	"github.com/yanking/price-watch/pkg/eventbus"
	"github.com/yanking/price-watch/internal/watch/exchange"
	"github.com/yanking/price-watch/pkg/log"
)

type Config struct {
	Log       log.Config                       `mapstructure:"log"`
	Redis     redisx.Config                    `mapstructure:"redis"`
	InfluxDB  influxdb.Config                  `mapstructure:"influxdb"`
	EventBus  eventbus.Config                  `mapstructure:"eventbus"`
	HTTP      exchange.HTTPConfig              `mapstructure:"http"`
	Exchanges map[string]exchange.ExchangeConfig `mapstructure:"exchanges"`
}
```

- [ ] **Step 3: Update config test**

Replace `internal/watch/config/config_test.go` content:

```go
package config

import (
	"testing"

	"github.com/yanking/price-watch/pkg/database/influxdb"
	"github.com/yanking/price-watch/pkg/database/redisx"
	"github.com/yanking/price-watch/pkg/eventbus"
	"github.com/yanking/price-watch/pkg/log"
)

func TestConfigStructure(t *testing.T) {
	c := Config{
		Log: log.Config{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		},
		Redis: redisx.Config{
			Addrs: []string{"localhost:6379"},
		},
		InfluxDB: influxdb.Config{
			URL:     "http://localhost:8086",
			Org:     "test",
			Buckets: influxdb.BucketsConfig{Tick: "prices-tick", Kline: "prices-kline"},
		},
		EventBus: eventbus.Config{
			Driver: "memory",
		},
	}

	if c.Log.Level != "info" {
		t.Errorf("Log.Level = %s, want info", c.Log.Level)
	}
	if len(c.Redis.Addrs) == 0 {
		t.Error("Redis.Addrs should not be empty")
	}
	if c.InfluxDB.URL == "" {
		t.Error("InfluxDB.URL should not be empty")
	}
	if c.EventBus.Driver != "memory" {
		t.Errorf("EventBus.Driver = %s, want memory", c.EventBus.Driver)
	}
}
```

- [ ] **Step 4: Update ServiceContext**

Replace `internal/watch/svc/serviceContext.go`. Remove MySQL, add InfluxDB and EventBus:

```go
package svc

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/yanking/price-watch/internal/watch/config"
	"github.com/yanking/price-watch/internal/watch/exchange"
	"github.com/yanking/price-watch/internal/watch/subscription"
	"github.com/yanking/price-watch/pkg/database/influxdb"
	"github.com/yanking/price-watch/pkg/database/redisx"
	"github.com/yanking/price-watch/pkg/eventbus"
	"github.com/yanking/price-watch/pkg/log"
)

type ServiceContext struct {
	Config   config.Config
	Logger   *slog.Logger
	Redis    *redisx.Client
	Influx   *influxdb.Client
	Bus      eventbus.Bus
	SubMgr   *subscription.Manager  // set in initial.App() after adapters are created
	Adapters []exchange.ExchangeAdapter // set in initial.App()
}

func NewServiceContext(c config.Config) (*ServiceContext, error) {
	if err := c.Log.Validate(); err != nil {
		return nil, fmt.Errorf("validate log config: %w", err)
	}

	logger, err := log.NewBuilder().FromConfig(&c.Log).Build()
	if err != nil {
		return nil, fmt.Errorf("create logger: %w", err)
	}

	redisClient, err := redisx.New(c.Redis)
	if err != nil {
		return nil, fmt.Errorf("create redis client: %w", err)
	}

	influxClient, err := influxdb.New(c.InfluxDB)
	if err != nil {
		return nil, fmt.Errorf("create influxdb client: %w", err)
	}

	bus, err := eventbus.New(c.EventBus, logger)
	if err != nil {
		return nil, fmt.Errorf("create eventbus: %w", err)
	}

	return &ServiceContext{
		Config: c,
		Logger: logger,
		Redis:  redisClient,
		Influx: influxClient,
		Bus:    bus,
		// SubMgr and Adapters are set later in initial.App()
	}, nil
}

// Close shuts down resources in reverse order per spec shutdown phases:
// Bus Drain → Bus Close → InfluxDB → Redis
// Note: Adapters are stopped via app.Server.Stop() before this is called.
func (ctx *ServiceContext) Close() error {
	var errs []error
	if ctx.Bus != nil {
		if err := ctx.Bus.Drain(); err != nil {
			errs = append(errs, fmt.Errorf("drain eventbus: %w", err))
		}
		if err := ctx.Bus.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close eventbus: %w", err))
		}
	}
	if ctx.Influx != nil {
		if err := ctx.Influx.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close influxdb: %w", err))
		}
	}
	if ctx.Redis != nil {
		if err := ctx.Redis.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close redis: %w", err))
		}
	}
	return errors.Join(errs...)
}
```

- [ ] **Step 5: Update ServiceContext test**

Replace `internal/watch/svc/serviceContext_test.go` to test with the new config (validation-only tests, no real connections):

```go
package svc

import (
	"strings"
	"testing"

	"github.com/yanking/price-watch/internal/watch/config"
	"github.com/yanking/price-watch/pkg/log"
)

func TestNewServiceContext_InvalidLogConfig(t *testing.T) {
	tests := []struct {
		name   string
		config config.Config
		errMsg string
	}{
		{
			name: "empty log level",
			config: config.Config{
				Log: log.Config{Level: "", Format: "json", Output: "stdout"},
			},
			errMsg: "level cannot be empty",
		},
		{
			name: "invalid format",
			config: config.Config{
				Log: log.Config{Level: "info", Format: "invalid", Output: "stdout"},
			},
			errMsg: "format must be",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewServiceContext(tt.config)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("error = %q, want containing %q", err, tt.errMsg)
			}
		})
	}
}
```

- [ ] **Step 6: Update watch.yaml**

Replace `configs/watch.yaml` with the full new config per spec.

- [ ] **Step 7: Run all tests**

```bash
go test ./internal/watch/... -v
```

Expected: all PASS

- [ ] **Step 8: Commit**

```bash
git add internal/watch/config/ internal/watch/svc/ internal/watch/exchange/config.go configs/watch.yaml
git commit -m "feat: update config and ServiceContext for price service (remove MySQL, add InfluxDB/EventBus)"
```

---

## Task 7: Exchange adapter — interface and BaseAdapter

**Files:**
- Create: `internal/watch/exchange/adapter.go`

- [ ] **Step 1: Write ExchangeAdapter interface and BaseAdapter**

Create `internal/watch/exchange/adapter.go`:

```go
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

	mu       sync.Mutex
	conn     *websocket.Conn
	ctx      context.Context
	cancel   context.CancelFunc
	symbols  []string
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
	// Remove from tracked symbols
	remaining := make([]string, 0, len(b.symbols))
	removeSet := make(map[string]bool, len(symbols))
	for _, s := range symbols {
		removeSet[s] = true
	}
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
	// Re-subscribe if we have symbols
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
			continue // skip unparseable messages (e.g., subscription confirmations)
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
```

- [ ] **Step 2: Write BaseAdapter unit tests**

Create `internal/watch/exchange/adapter_test.go`:

```go
package exchange

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/yanking/price-watch/internal/watch/event"
)

func TestPublishParsed_TickEvent(t *testing.T) {
	// Use in-memory bus to verify publishParsed routes tick events correctly
	bus, _ := eventbus.newMemoryBus(eventbus.MemoryConfig{BufferSize: 8})
	defer bus.Close()

	var received []byte
	done := make(chan struct{})
	bus.Subscribe("price.tick.binance.BTCUSDT", func(subject string, data []byte) error {
		received = append([]byte{}, data...)
		close(done)
		return nil
	})

	b := &BaseAdapter{name: "binance", bus: bus, logger: slog.Default()}
	b.ctx, b.cancel = context.WithCancel(context.Background())

	tick := event.NewTickEvent("binance", "BTCUSDT", event.TickData{
		Price:  decimal.NewFromFloat(67000.50),
		Volume: decimal.NewFromFloat(1.5),
	})
	b.publishParsed(tick)

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timed out")
	}

	if len(received) == 0 {
		t.Fatal("no data received")
	}
}

func TestReconnectDelay(t *testing.T) {
	// Verify exponential backoff capped by ReconnectMax
	cfg := ExchangeConfig{
		ReconnectBase: time.Second,
		ReconnectMax:  10 * time.Second,
	}

	// attempt 0: 1s, attempt 1: 2s, attempt 3: 8s, attempt 4: 10s (capped)
	delays := []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second, 8 * time.Second, 10 * time.Second}
	for i, want := range delays {
		got := time.Duration(math.Min(
			float64(cfg.ReconnectBase)*math.Pow(2, float64(i)),
			float64(cfg.ReconnectMax),
		))
		if got != want {
			t.Errorf("attempt %d: got %v, want %v", i, got, want)
		}
	}
}
```

Note: These tests validate event publishing and reconnect delay calculation without needing a real WebSocket server. Integration tests for full WebSocket flow are deferred.

- [ ] **Step 3: Run tests**

```bash
go test ./internal/watch/exchange/... -v
```

Expected: PASS

Also add the adapter factory function used by `initApp.go`:

```go
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
	case "gateio":
		return NewGateAdapter(bus, logger, cfg)
	default:
		return nil
	}
}
```

Note: `NewBinanceAdapter` etc. are stubs until Tasks 8-11. Initially they can return nil or panic — they will be implemented in subsequent tasks.

- [ ] **Step 4: Commit**

```bash
git add internal/watch/exchange/adapter.go internal/watch/exchange/adapter_test.go
git commit -m "feat: add ExchangeAdapter interface and BaseAdapter with reconnect/ping"
```

---

## Task 8: Binance exchange adapter

**Files:**
- Create: `internal/watch/exchange/binance.go`
- Create: `internal/watch/exchange/binance_test.go`

- [ ] **Step 1: Write failing tests for Binance parser and builder**

Create `internal/watch/exchange/binance_test.go` testing:
- `parseBinanceMessage` correctly parses a trade message into `Event[TickData]`
- `buildBinanceSubscribeMsg` produces correct JSON
- `buildBinanceUnsubscribeMsg` produces correct JSON
- Symbol normalization (Binance symbols are already canonical)

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/watch/exchange/... -run TestBinance -v
```

- [ ] **Step 3: Write Binance adapter**

Create `internal/watch/exchange/binance.go` implementing:
- `NewBinanceAdapter(bus, logger, cfg) *BinanceAdapter`
- `parseBinanceMessage(msg []byte) (any, error)` — parse `{"e":"trade","s":"BTCUSDT","p":"67000.50","q":"1.5","T":1234567890}` format
- `buildBinanceSubscribeMsg(symbols) ([]byte, error)` — `{"method":"SUBSCRIBE","params":["btcusdt@trade"],"id":1}`
- `buildBinanceUnsubscribeMsg(symbols) ([]byte, error)`
- `FetchKlines(ctx, symbol, interval, start, end) ([]KlineData, error)` — REST GET `/api/v3/klines` with rate limiter

- [ ] **Step 4: Run tests**

```bash
go test ./internal/watch/exchange/... -run TestBinance -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/watch/exchange/binance.go internal/watch/exchange/binance_test.go
git commit -m "feat: add Binance exchange adapter"
```

---

## Task 9: OKX exchange adapter

**Files:**
- Create: `internal/watch/exchange/okx.go`
- Create: `internal/watch/exchange/okx_test.go`

**OKX specifics:**

Symbol normalization: `BTC-USDT` → `BTCUSDT` (remove `-`), reverse: `BTCUSDT` → `BTC-USDT` (insert `-` before `USDT`/`USDC`/`USD`)

Subscribe message:
```json
{"op":"subscribe","args":[{"channel":"tickers","instId":"BTC-USDT"}]}
```

Unsubscribe message:
```json
{"op":"unsubscribe","args":[{"channel":"tickers","instId":"BTC-USDT"}]}
```

Incoming ticker message:
```json
{"arg":{"channel":"tickers","instId":"BTC-USDT"},"data":[{"instId":"BTC-USDT","last":"67000.5","vol24h":"1500.5","ts":"1711360000000"}]}
```

REST K-lines: `GET /api/v5/market/candles?instId=BTC-USDT&bar=1m&before={startMs}&after={endMs}&limit=300`

Response: `{"code":"0","data":[["1711360000000","67000","67100","66950","67050","120.5",...]]}`

Array indices: [0]=ts, [1]=open, [2]=high, [3]=low, [4]=close, [5]=volume

Ping: send `"ping"` text, expect `"pong"` response

- [ ] **Step 1: Write failing tests**

Test `parseOKXMessage`, `buildOKXSubscribeMsg`, `normalizeOKXSymbol`, `denormalizeOKXSymbol`

- [ ] **Step 2: Run tests to verify they fail**
- [ ] **Step 3: Write OKX adapter implementation**
- [ ] **Step 4: Run tests, verify PASS**
- [ ] **Step 5: Commit**

```bash
git commit -m "feat: add OKX exchange adapter"
```

---

## Task 10: Bybit exchange adapter

**Files:**
- Create: `internal/watch/exchange/bybit.go`
- Create: `internal/watch/exchange/bybit_test.go`

**Bybit specifics:**

Symbol normalization: already canonical (Bybit uses `BTCUSDT`)

Subscribe message:
```json
{"op":"subscribe","args":["tickers.BTCUSDT"]}
```

Unsubscribe message:
```json
{"op":"unsubscribe","args":["tickers.BTCUSDT"]}
```

Incoming ticker message:
```json
{"topic":"tickers.BTCUSDT","type":"snapshot","data":{"symbol":"BTCUSDT","lastPrice":"67000.5","volume24h":"1500.5","turnover24h":"100500000"}}
```

REST K-lines: `GET /v5/market/kline?category=spot&symbol=BTCUSDT&interval=1&start={startMs}&end={endMs}&limit=200`

Response: `{"retCode":0,"result":{"list":[["1711360000000","67000","67100","66950","67050","120.5","8070000"]]}}`

Array indices: [0]=ts, [1]=open, [2]=high, [3]=low, [4]=close, [5]=volume, [6]=turnover

Interval mapping: 1m→"1", 5m→"5", 15m→"15", 1h→"60", 4h→"240", 1d→"D"

Ping: `{"op":"ping"}`, expect `{"op":"pong",...}`

- [ ] **Step 1: Write failing tests**

Test `parseBybitMessage`, `buildBybitSubscribeMsg`, `mapBybitInterval`

- [ ] **Step 2: Run tests to verify they fail**
- [ ] **Step 3: Write Bybit adapter implementation**
- [ ] **Step 4: Run tests, verify PASS**
- [ ] **Step 5: Commit**

```bash
git commit -m "feat: add Bybit exchange adapter"
```

---

## Task 11: Gate.io exchange adapter

**Files:**
- Create: `internal/watch/exchange/gateio.go`
- Create: `internal/watch/exchange/gateio_test.go`

**Gate.io specifics:**

Symbol normalization: `BTC_USDT` → `BTCUSDT` (remove `_`), reverse: `BTCUSDT` → `BTC_USDT` (insert `_` before `USDT`/`USDC`/`USD`)

Subscribe message:
```json
{"time":1711360000,"channel":"spot.tickers","event":"subscribe","payload":["BTC_USDT"]}
```

Unsubscribe message:
```json
{"time":1711360000,"channel":"spot.tickers","event":"unsubscribe","payload":["BTC_USDT"]}
```

Incoming ticker message:
```json
{"time":1711360000,"channel":"spot.tickers","event":"update","result":{"currency_pair":"BTC_USDT","last":"67000.5","base_volume":"1500.5"}}
```

REST K-lines: `GET /api/v4/spot/candlesticks?currency_pair=BTC_USDT&interval=1m&from={startUnix}&to={endUnix}&limit=1000`

Response: `[["1711360000","120.5","67050","67100","66950","67000",...]]`

Array indices: [0]=ts(unix), [1]=volume, [2]=close, [3]=high, [4]=low, [5]=open

Ping: `{"time":1711360000,"channel":"spot.ping"}`, expect pong response

- [ ] **Step 1: Write failing tests**

Test `parseGateMessage`, `buildGateSubscribeMsg`, `normalizeGateSymbol`, `denormalizeGateSymbol`

- [ ] **Step 2: Run tests to verify they fail**
- [ ] **Step 3: Write Gate.io adapter implementation**
- [ ] **Step 4: Run tests, verify PASS**
- [ ] **Step 5: Commit**

```bash
git commit -m "feat: add Gate.io exchange adapter"
```

---

## Task 12: SubscriptionManager

**Files:**
- Create: `internal/watch/subscription/manager.go`
- Create: `internal/watch/subscription/manager_test.go`

- [ ] **Step 1: Write failing tests**

Create `internal/watch/subscription/manager_test.go` using `alicebob/miniredis` (already a dependency). Test:
- `Add` persists symbols to Redis set `sub:{exchange}`
- `Add` with empty exchanges targets all adapters
- `Remove` removes symbols from Redis
- `List` returns all subscriptions
- `Restore` reads Redis and calls Subscribe on adapters
- Partial failure: one adapter fails, others succeed, result reflects this

Use a mock adapter (simple struct implementing ExchangeAdapter with controllable error).

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/watch/subscription/... -v
```

- [ ] **Step 3: Write SubscriptionManager implementation**

Create `internal/watch/subscription/manager.go`:

```go
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
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/watch/subscription/... -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/watch/subscription/
git commit -m "feat: add SubscriptionManager with Redis persistence and partial failure handling"
```

---

## Task 13: CacheSubscriber

**Files:**
- Create: `internal/watch/consumer/cache.go`
- Create: `internal/watch/consumer/cache_test.go`

- [ ] **Step 1: Write failing tests**

Test with miniredis + in-memory eventbus:
- Publish a TickData event → verify Redis Hash `price:{exchange}:{symbol}` contains correct data
- Publish events from two exchanges → verify Sorted Set `price:all:{symbol}` has both members
- TTL is set to 5 minutes

- [ ] **Step 2: Run tests to verify they fail**
- [ ] **Step 3: Write CacheSubscriber**

```go
package consumer

// CacheSubscriber subscribes to price.tick.> and writes to Redis.
// On Redis failure: log warning and continue (cache is ephemeral).
type CacheSubscriber struct { ... }

func NewCacheSubscriber(rds redis.Cmdable, bus eventbus.Bus, logger *slog.Logger) *CacheSubscriber
func (c *CacheSubscriber) Start() error   // subscribes to bus
func (c *CacheSubscriber) Stop() error    // unsubscribes
```

- [ ] **Step 4: Run tests, verify PASS**
- [ ] **Step 5: Commit**

```bash
git commit -m "feat: add CacheSubscriber writing tick prices to Redis"
```

---

## Task 14: StorageSubscriber

**Files:**
- Create: `internal/watch/consumer/storage.go`
- Create: `internal/watch/consumer/storage_test.go`

- [ ] **Step 1: Write failing tests**

Test with in-memory eventbus + mock InfluxDB write API:
- Publish tick events → verify points are batched and written
- Publish kline events → verify written to kline bucket
- Flush on interval (use short interval for test)

- [ ] **Step 2: Run tests to verify they fail**
- [ ] **Step 3: Write StorageSubscriber**

```go
package consumer

// StorageSubscriber subscribes to price.tick.> and price.kline.>,
// batches writes to InfluxDB with configurable batch size and flush interval.
type StorageSubscriber struct { ... }

func NewStorageSubscriber(influx *influxdb.Client, bus eventbus.Bus, logger *slog.Logger, batchSize int, flushInterval time.Duration) *StorageSubscriber
func (s *StorageSubscriber) Start() error
func (s *StorageSubscriber) Stop() error  // flush remaining, then stop
```

- [ ] **Step 4: Run tests, verify PASS**
- [ ] **Step 5: Commit**

```bash
git commit -m "feat: add StorageSubscriber writing to InfluxDB with batching"
```

---

## Task 15a: GinServer + response helper + health endpoints

**Files:**
- Create: `internal/watch/server/http.go`
- Create: `internal/watch/handler/response.go`
- Create: `internal/watch/handler/health.go`
- Create: `internal/watch/handler/routes.go`
- Create: `internal/watch/handler/health_test.go`

- [ ] **Step 1: Write GinServer**

Create `internal/watch/server/http.go`:

```go
package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yanking/price-watch/internal/watch/svc"
)

type GinServer struct {
	engine *gin.Engine
	server *http.Server
}

func NewGinServer(ctx *svc.ServiceContext, setupRoutes func(*gin.Engine)) *GinServer {
	engine := gin.New()
	engine.Use(gin.Recovery())
	setupRoutes(engine)

	return &GinServer{
		engine: engine,
		server: &http.Server{
			Addr:    ctx.Config.HTTP.Addr,
			Handler: engine,
		},
	}
}

func (s *GinServer) Start() error {
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("http server: %w", err)
	}
	return nil
}

func (s *GinServer) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return s.server.Shutdown(ctx)
}

func (s *GinServer) String() string { return "http-server" }
```

- [ ] **Step 2: Write response helper and route registration**

Create `internal/watch/handler/response.go`:

```go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{Code: 0, Message: "ok", Data: data})
}

func Error(c *gin.Context, httpCode int, msg string) {
	c.JSON(httpCode, Response{Code: -1, Message: msg})
}
```

Create `internal/watch/handler/routes.go`:

```go
package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/yanking/price-watch/internal/watch/svc"
)

func RegisterRoutes(e *gin.Engine, ctx *svc.ServiceContext) {
	h := NewHealthHandler(ctx)
	e.GET("/health", h.Health)
	e.GET("/ready", h.Ready)

	v1 := e.Group("/api/v1")
	{
		sub := NewSubscriptionHandler(ctx)
		v1.POST("/subscriptions", sub.Add)
		v1.DELETE("/subscriptions", sub.Remove)
		v1.GET("/subscriptions", sub.List)

		price := NewPriceHandler(ctx)
		v1.GET("/prices", price.ListAll)
		v1.GET("/prices/:symbol", price.GetBySymbol)

		kline := NewKlineHandler(ctx)
		v1.GET("/klines/:symbol", kline.Get)
	}
}
```

- [ ] **Step 3: Write health handler**

Create `internal/watch/handler/health.go`:

```go
package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yanking/price-watch/internal/watch/svc"
)

type HealthHandler struct {
	ctx *svc.ServiceContext
}

func NewHealthHandler(ctx *svc.ServiceContext) *HealthHandler {
	return &HealthHandler{ctx: ctx}
}

func (h *HealthHandler) Health(c *gin.Context) {
	OK(c, gin.H{"status": "alive"})
}

func (h *HealthHandler) Ready(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	if err := h.ctx.Redis.Ping(ctx); err != nil {
		Error(c, http.StatusServiceUnavailable, "redis not ready")
		return
	}
	if err := h.ctx.Influx.Ping(ctx); err != nil {
		Error(c, http.StatusServiceUnavailable, "influxdb not ready")
		return
	}
	hasAdapter := len(h.ctx.Adapters) > 0
	if !hasAdapter {
		Error(c, http.StatusServiceUnavailable, "no exchange adapters")
		return
	}
	OK(c, gin.H{"status": "ready"})
}
```

- [ ] **Step 4: Write health tests**

Create `internal/watch/handler/health_test.go` using `httptest.NewRecorder`. Test `/health` returns 200 with `{"code":0}`.

- [ ] **Step 5: Run tests**

```bash
go test ./internal/watch/handler/... -v
go test ./internal/watch/server/... -v
```

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/watch/server/ internal/watch/handler/response.go internal/watch/handler/routes.go internal/watch/handler/health.go internal/watch/handler/health_test.go
git commit -m "feat: add GinServer, response helpers, health endpoints, and route registration"
```

---

## Task 15b: Subscription handlers

**Files:**
- Create: `internal/watch/handler/subscription.go`
- Create: `internal/watch/handler/subscription_test.go`

- [ ] **Step 1: Write subscription handler**

Create `internal/watch/handler/subscription.go`:

```go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yanking/price-watch/internal/watch/subscription"
	"github.com/yanking/price-watch/internal/watch/svc"
)

type SubscriptionHandler struct {
	ctx *svc.ServiceContext
}

func NewSubscriptionHandler(ctx *svc.ServiceContext) *SubscriptionHandler {
	return &SubscriptionHandler{ctx: ctx}
}

func (h *SubscriptionHandler) Add(c *gin.Context) {
	var req subscription.SubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if len(req.Symbols) == 0 {
		Error(c, http.StatusBadRequest, "symbols is required")
		return
	}
	result, err := h.ctx.SubMgr.Add(c.Request.Context(), req)
	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	OK(c, result)
}

func (h *SubscriptionHandler) Remove(c *gin.Context) {
	var req subscription.UnsubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := h.ctx.SubMgr.Remove(c.Request.Context(), req)
	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	OK(c, result)
}

func (h *SubscriptionHandler) List(c *gin.Context) {
	infos, err := h.ctx.SubMgr.List(c.Request.Context())
	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	OK(c, infos)
}
```

- [ ] **Step 2: Write subscription handler tests**

Create `internal/watch/handler/subscription_test.go` using `httptest.NewRecorder`. Test POST returns correct response format with mock SubMgr.

- [ ] **Step 3: Run tests**

```bash
go test ./internal/watch/handler/... -run TestSubscription -v
```

Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/watch/handler/subscription.go internal/watch/handler/subscription_test.go
git commit -m "feat: add subscription CRUD handlers"
```

---

## Task 15c: Price and K-line handlers

**Files:**
- Create: `internal/watch/handler/price.go`
- Create: `internal/watch/handler/kline.go`
- Create: `internal/watch/handler/price_test.go`

- [ ] **Step 1: Write price handler**

Create `internal/watch/handler/price.go`:

```go
package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yanking/price-watch/internal/watch/svc"
)

type PriceHandler struct {
	ctx *svc.ServiceContext
}

func NewPriceHandler(ctx *svc.ServiceContext) *PriceHandler {
	return &PriceHandler{ctx: ctx}
}

// GetBySymbol returns latest prices across all exchanges for a symbol.
// Reads from Redis Sorted Set `price:all:{symbol}` and Hash `price:{exchange}:{symbol}`.
func (h *PriceHandler) GetBySymbol(c *gin.Context) {
	symbol := c.Param("symbol")
	rds := h.ctx.Redis.Client()

	// Get all exchange prices from sorted set
	key := fmt.Sprintf("price:all:%s", symbol)
	members, err := rds.ZRangeWithScores(c.Request.Context(), key, 0, -1).Result()
	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	type ExchangePrice struct {
		Exchange string `json:"exchange"`
		Price    string `json:"price"`
	}
	prices := make([]ExchangePrice, 0, len(members))
	for _, m := range members {
		prices = append(prices, ExchangePrice{
			Exchange: m.Member.(string),
			Price:    fmt.Sprintf("%g", m.Score),
		})
	}
	OK(c, gin.H{"symbol": symbol, "prices": prices})
}

// ListAll returns latest prices for all subscribed symbols.
func (h *PriceHandler) ListAll(c *gin.Context) {
	// Scan for all price:all:* keys
	rds := h.ctx.Redis.Client()
	var cursor uint64
	var allPrices []gin.H

	for {
		keys, nextCursor, err := rds.Scan(c.Request.Context(), cursor, "price:all:*", 100).Result()
		if err != nil {
			Error(c, http.StatusInternalServerError, err.Error())
			return
		}
		for _, key := range keys {
			symbol := key[len("price:all:"):]
			members, _ := rds.ZRangeWithScores(c.Request.Context(), key, 0, -1).Result()
			exchanges := make([]gin.H, 0, len(members))
			for _, m := range members {
				exchanges = append(exchanges, gin.H{
					"exchange": m.Member, "price": fmt.Sprintf("%g", m.Score),
				})
			}
			allPrices = append(allPrices, gin.H{"symbol": symbol, "prices": exchanges})
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	OK(c, allPrices)
}
```

- [ ] **Step 2: Write kline handler**

Create `internal/watch/handler/kline.go`:

```go
package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yanking/price-watch/internal/watch/svc"
)

type KlineHandler struct {
	ctx *svc.ServiceContext
}

func NewKlineHandler(ctx *svc.ServiceContext) *KlineHandler {
	return &KlineHandler{ctx: ctx}
}

// Get queries historical K-lines from InfluxDB.
// Query params: exchange, interval (default 1m), start, end, limit (default 500)
// Falls back to adapter REST if InfluxDB has no data.
func (h *KlineHandler) Get(c *gin.Context) {
	symbol := c.Param("symbol")
	exc := c.Query("exchange")
	interval := c.DefaultQuery("interval", "1m")
	limitStr := c.DefaultQuery("limit", "500")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 || limit > 5000 {
		limit = 500
	}

	startStr := c.Query("start")
	endStr := c.Query("end")
	end := time.Now()
	start := end.Add(-24 * time.Hour)
	if startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			start = t
		}
	}
	if endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			end = t
		}
	}

	bucket := h.ctx.Influx.Config().Buckets.Kline
	// Build Flux query
	filter := fmt.Sprintf(`|> filter(fn: (r) => r["symbol"] == "%s")`, symbol)
	if exc != "" {
		filter += fmt.Sprintf(` |> filter(fn: (r) => r["exchange"] == "%s")`, exc)
	}
	query := fmt.Sprintf(`from(bucket: "%s")
		|> range(start: %s, stop: %s)
		|> filter(fn: (r) => r["_measurement"] == "spot_kline")
		|> filter(fn: (r) => r["interval"] == "%s")
		%s
		|> limit(n: %d)`,
		bucket, start.Format(time.RFC3339), end.Format(time.RFC3339), interval, filter, limit)

	result, err := h.ctx.Influx.Query(c.Request.Context(), query)
	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	var klines []gin.H
	for result.Next() {
		record := result.Record()
		klines = append(klines, gin.H{
			"time":   record.Time(),
			"open":   record.ValueByKey("open"),
			"high":   record.ValueByKey("high"),
			"low":    record.ValueByKey("low"),
			"close":  record.ValueByKey("close"),
			"volume": record.ValueByKey("volume"),
		})
	}

	// Fallback to adapter REST if no data in InfluxDB
	if len(klines) == 0 && exc != "" {
		for _, adapter := range h.ctx.Adapters {
			if adapter.Name() == exc {
				data, fErr := adapter.FetchKlines(c.Request.Context(), symbol, interval, start, end)
				if fErr != nil {
					Error(c, http.StatusInternalServerError, fErr.Error())
					return
				}
				for _, k := range data {
					klines = append(klines, gin.H{
						"interval": k.Interval,
						"open":     k.Open.String(),
						"high":     k.High.String(),
						"low":      k.Low.String(),
						"close":    k.Close.String(),
						"volume":   k.Volume.String(),
					})
				}
				break
			}
		}
	}

	OK(c, gin.H{"symbol": symbol, "interval": interval, "klines": klines})
}
```

- [ ] **Step 3: Write price handler tests**

Create `internal/watch/handler/price_test.go` using `httptest.NewRecorder` + miniredis. Seed Redis with test price data, verify GET `/api/v1/prices/BTCUSDT` returns correct structure.

- [ ] **Step 4: Run tests**

```bash
go test ./internal/watch/handler/... -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/watch/handler/price.go internal/watch/handler/kline.go internal/watch/handler/price_test.go
git commit -m "feat: add price and kline query handlers"
```

---

## Task 16: Wire everything together — initial/ bootstrap

**Files:**
- Modify: `cmd/watch/initial/initApp.go`
- Modify: `cmd/watch/initial/createService.go`
- Modify: `cmd/watch/initial/close.go`
- Modify: `cmd/watch/main.go`

Shutdown ordering per spec: Adapters stop (via Server.Stop) → Bus Drain + Close → InfluxDB close → Redis close. The `ctx.Close()` in `main.go` handles phases 2-5 (Bus, InfluxDB, Redis). Server stops are handled by the app framework calling `close.go` cleanup functions. These two do not overlap — `close.go` only stops servers, `ctx.Close()` only closes infrastructure.

- [ ] **Step 1: Update initApp.go**

Create exchange adapters, subscription manager, start consumers, restore subscriptions:

```go
package initial

import (
	"context"

	"github.com/yanking/price-watch/internal/watch/consumer"
	"github.com/yanking/price-watch/internal/watch/exchange"
	"github.com/yanking/price-watch/internal/watch/subscription"
	"github.com/yanking/price-watch/internal/watch/svc"
)

func App(ctx *svc.ServiceContext) {
	// Create exchange adapters from config
	adapterMap := make(map[string]exchange.ExchangeAdapter)
	for name, cfg := range ctx.Config.Exchanges {
		if !cfg.Enabled {
			continue
		}
		adapter := exchange.NewAdapter(name, ctx.Bus, ctx.Logger, cfg)
		if adapter != nil {
			adapterMap[name] = adapter
			ctx.Adapters = append(ctx.Adapters, adapter)
		}
	}

	// Create SubscriptionManager
	ctx.SubMgr = subscription.NewManager(ctx.Redis.Client(), adapterMap, ctx.Logger)

	// Start CacheSubscriber
	cacheSub := consumer.NewCacheSubscriber(ctx.Redis.Client(), ctx.Bus, ctx.Logger)
	if err := cacheSub.Start(); err != nil {
		ctx.Logger.Error("start cache subscriber", "error", err)
	}

	// Start StorageSubscriber
	storageSub := consumer.NewStorageSubscriber(
		ctx.Influx, ctx.Bus, ctx.Logger,
		int(ctx.Config.InfluxDB.BatchSize),
		ctx.Config.InfluxDB.FlushInterval,
	)
	if err := storageSub.Start(); err != nil {
		ctx.Logger.Error("start storage subscriber", "error", err)
	}

	// Restore subscriptions from Redis
	if err := ctx.SubMgr.Restore(context.Background()); err != nil {
		ctx.Logger.Error("restore subscriptions", "error", err)
	}
}
```

- [ ] **Step 2: Update createService.go**

Register HTTP server and exchange adapter wrappers:

```go
package initial

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/yanking/price-watch/internal/watch/exchange"
	"github.com/yanking/price-watch/internal/watch/handler"
	"github.com/yanking/price-watch/internal/watch/server"
	"github.com/yanking/price-watch/internal/watch/svc"
	"github.com/yanking/price-watch/pkg/app"
)

func CreateServices(ctx *svc.ServiceContext) (services []app.Server) {
	// HTTP Server
	httpSrv := server.NewGinServer(ctx, func(e *gin.Engine) {
		handler.RegisterRoutes(e, ctx)
	})
	services = append(services, httpSrv)

	// Exchange Adapter wrappers (each adapter as a Server)
	for _, adapter := range ctx.Adapters {
		services = append(services, &adapterServer{adapter: adapter})
	}

	return
}

type adapterServer struct {
	adapter exchange.ExchangeAdapter
}

func (s *adapterServer) Start() error   { return s.adapter.Start(context.Background()) }
func (s *adapterServer) Stop() error    { return s.adapter.Stop() }
func (s *adapterServer) String() string { return fmt.Sprintf("exchange-%s", s.adapter.Name()) }
```

- [ ] **Step 3: Update close.go**

Server stops only — infrastructure cleanup is handled by `ctx.Close()` in `main.go`:

```go
package initial

import "github.com/yanking/price-watch/pkg/app"

func Close(servers []app.Server) (closes []app.CleanupFunc) {
	// Stop servers in reverse order: adapters first (registered last), then HTTP
	for i := len(servers) - 1; i >= 0; i-- {
		s := servers[i]
		closes = append(closes, s.Stop)
	}
	return
}
```

Note: `close.go` signature stays as `Close(servers []app.Server)` — same as the original. No change to `main.go` call site. The `defer ctx.Close()` in `main.go` handles Bus Drain/Close, InfluxDB close, and Redis close after `app.Run()` returns.

- [ ] **Step 4: Verify main.go is consistent**

The existing `main.go` already has:
- `closes := initial.Close(servers)` — stops servers
- `defer ctx.Close()` — closes Bus/InfluxDB/Redis

Shutdown sequence: signal → `app.stop()` calls `Close()` cleanups (servers stop) → `app.Run()` returns → `defer ctx.Close()` fires (Bus drain/close, InfluxDB close, Redis close). This matches the spec's 5-phase shutdown ordering.

No changes needed to `main.go`.

- [ ] **Step 5: Verify compilation**

```bash
go build ./cmd/watch/...
```

Expected: no errors

- [ ] **Step 6: Commit**

```bash
git add cmd/watch/
git commit -m "feat: wire bootstrap — register HTTP server, adapters, consumers, and ordered cleanup"
```

---

## Task 17: Docker infrastructure updates

**Files:**
- Modify: `docker-compose.yml`
- Modify: `.env.example`

- [ ] **Step 1: Add InfluxDB and NATS to docker-compose.yml**

Add after the redis service:

```yaml
  influxdb:
    image: influxdb:${INFLUXDB_VERSION:-2.7-alpine}
    restart: unless-stopped
    ports:
      - "${INFLUXDB_PORT:-8086}:8086"
    volumes:
      - ${DATA_DIR:-./docker/data}/influxdb:/var/lib/influxdb2
    environment:
      DOCKER_INFLUXDB_INIT_MODE: setup
      DOCKER_INFLUXDB_INIT_USERNAME: admin
      DOCKER_INFLUXDB_INIT_PASSWORD: ${INFLUXDB_PASSWORD:-adminadmin}
      DOCKER_INFLUXDB_INIT_ORG: ${INFLUXDB_ORG:-price-watch}
      DOCKER_INFLUXDB_INIT_BUCKET: prices-tick
    healthcheck:
      test: ["CMD", "influx", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - price-watch-net

  nats:
    image: nats:${NATS_VERSION:-2.10-alpine}
    restart: unless-stopped
    command: ["--jetstream", "--store_dir", "/data"]
    ports:
      - "${NATS_PORT:-4222}:4222"
      - "${NATS_MONITOR_PORT:-8222}:8222"
    volumes:
      - ${DATA_DIR:-./docker/data}/nats:/data
    healthcheck:
      test: ["CMD-SHELL", "wget -q --spider http://localhost:8222/healthz || exit 1"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - price-watch-net
```

- [ ] **Step 2: Add env vars to .env.example**

```
# ==================== InfluxDB ====================

INFLUXDB_VERSION=2.7-alpine
INFLUXDB_PORT=8086
INFLUXDB_PASSWORD=adminadmin
INFLUXDB_ORG=price-watch

# ==================== NATS ====================

NATS_VERSION=2.10-alpine
NATS_PORT=4222
NATS_MONITOR_PORT=8222
```

- [ ] **Step 3: Commit**

```bash
git add docker-compose.yml .env.example
git commit -m "feat: add InfluxDB and NATS to Docker infrastructure"
```

---

## Task 18: Full integration test

**Files:**
- Create: `cmd/watch/initial/integration_test.go` (optional, tag-gated)

- [ ] **Step 1: Run all unit tests**

```bash
make test
```

Expected: all PASS

- [ ] **Step 2: Run linter**

```bash
make lint
```

Expected: no errors (or only pre-existing warnings)

- [ ] **Step 3: Verify build**

```bash
make build
```

Expected: binary produced in `bin/watch`

- [ ] **Step 4: Verify config loading**

```bash
./bin/watch -v
```

Expected: version info printed

- [ ] **Step 5: Commit any fixes**

```bash
git add -A
git commit -m "fix: address issues found during integration verification"
```

(Only if there are fixes needed.)

---

## Dependency Graph

```
Task 1 (deps)
  ├→ Task 2 (eventbus memory) → Task 3 (eventbus NATS)
  ├→ Task 4 (influxdb client)
  ├→ Task 5 (event model)
  └→ Task 6 (config + ServiceContext) ← depends on Tasks 2, 4, 5
      ├→ Task 7 (BaseAdapter) ← depends on Tasks 2, 5
      │   ├→ Task 8 (Binance)  ┐
      │   ├→ Task 9 (OKX)      ├ parallel
      │   ├→ Task 10 (Bybit)   │
      │   └→ Task 11 (Gate.io) ┘
      ├→ Task 12 (SubscriptionManager) ← depends on Task 7
      ├→ Task 13 (CacheSubscriber) ← depends on Tasks 2, 5
      ├→ Task 14 (StorageSubscriber) ← depends on Tasks 2, 4, 5
      ├→ Task 15a (GinServer + health) ← depends on Task 6
      ├→ Task 15b (Subscription handlers) ← depends on Tasks 12, 15a
      ├→ Task 15c (Price + kline handlers) ← depends on Tasks 13, 14, 15a
      └→ Task 16 (Bootstrap wiring) ← depends on all above
          ├→ Task 17 (Docker) — independent, can run anytime
          └→ Task 18 (Integration verification)
```

**Parallelizable groups after Task 1:**
- Tasks 2, 4, 5 (no interdependencies)
- Tasks 8, 9, 10, 11 (all depend on Task 7 only)
- Tasks 12, 13, 14 (independent consumers/manager)
- Task 17 (Docker) can run at any point
