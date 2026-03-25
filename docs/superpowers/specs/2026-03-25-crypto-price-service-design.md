# Crypto Price Fetching Service Design

## Overview

Design for a cryptocurrency spot price fetching service within the price-watch system. The service connects to multiple exchanges via WebSocket for real-time price data, fetches historical K-line data via REST, and distributes events through an internal event bus to cache and time-series storage consumers.

## Key Decisions

| Decision | Choice | Reasoning |
|----------|--------|-----------|
| Data source | Exchange direct APIs (Binance, OKX, Bybit, Gate.io) | Real-time spot trading data |
| Real-time protocol | WebSocket | Low-latency, exchange-pushed price updates |
| Historical data | REST API | Standard for K-line endpoints |
| Real-time cache | Redis | Microsecond reads for latest prices |
| Historical storage | InfluxDB (two buckets) | Purpose-built for time-series data |
| Event distribution | EventBus (pluggable: in-memory / NATS JetStream) | Decoupled consumers, configurable reliability |
| HTTP framework | Gin | Mature ecosystem, middleware-rich, high performance |
| Subscription management | API-driven, persisted to Redis | Runtime flexibility without restart |
| K-line intervals | 1m, 5m, 15m, 1h, 4h, 1d | Standard intervals supported by all four exchanges |
| Alert notifications | Deferred to future iteration | Focus on data acquisition and storage first |
| MySQL | Not used in this phase | No relational data needs yet; `pkg/database/mysqlx/` preserved for future use |
| Serialization | JSON throughout | Simple, debuggable, sufficient performance for current scale |
| Event ID | ULID | Sortable, unique, no coordination needed |
| Symbol format | Uppercase no separator (e.g. `BTCUSDT`) | Canonical format; each adapter normalizes from exchange-native format |

## Architecture

### Event-Driven with EventBus

```
Exchange Adapters ──→ EventBus ──→ CacheSubscriber (Redis)
                             ──→ StorageSubscriber (InfluxDB)
                             ──→ (future: AlertSubscriber)
```

All exchange adapters publish unified events to the EventBus. Subscribers consume events independently. Adding new consumers (e.g., alerts) requires only subscribing to the bus — no changes to adapters or existing consumers.

### Shutdown Ordering

Shutdown proceeds in phases to ensure data integrity:

1. **Exchange Adapters** stop first — close WebSocket connections, no new events produced
2. **EventBus** calls `Drain()` to flush remaining messages to subscribers, then `Close()`
3. **StorageSubscriber** flushes pending InfluxDB batch, then stops
4. **CacheSubscriber** stops (no flush needed, cache is ephemeral)
5. **InfluxDB / Redis clients** close connections

The `App` cleanup functions are registered in this order. Exchange adapters implement `app.Server` via a thin wrapper that bridges `Start(ctx)` to the no-arg `Server.Start()` by capturing the app-level context at construction time.

## Event Model

### Unified Envelope (CloudEvents-inspired, Go generics)

```go
type Event[T any] struct {
    ID        string    `json:"id"`        // ULID, globally unique, sortable
    Type      string    `json:"type"`      // "price.tick", "price.kline"
    Source    string    `json:"source"`    // "binance", "okx", "bybit", "gateio"
    Subject   string    `json:"subject"`   // "BTCUSDT" (canonical format)
    Timestamp time.Time `json:"timestamp"`
    Data      T         `json:"data"`
}

// BuildSubject derives the EventBus subject from envelope fields.
// Single source of truth — adapters and publishers use this, never construct subjects manually.
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
```

- Envelope handles routing (Type, Source for filtering), payload carries domain data
- `decimal.Decimal` for price fields to avoid floating-point precision issues
- `BuildSubject()` ensures EventBus subject and envelope fields never diverge

### Serialization Strategy

All events are serialized as JSON. The business layer provides helper functions:

```go
// Marshal serializes a typed event to []byte for the EventBus
func Marshal[T any](e Event[T]) ([]byte, error)

// Unmarshal deserializes from []byte. Two-pass approach:
// 1. Unmarshal to Event[json.RawMessage] to inspect Type field
// 2. Unmarshal Data to the concrete type based on Type
func Unmarshal(data []byte) (eventType string, raw Event[json.RawMessage], err error)
func UnmarshalData[T any](raw json.RawMessage) (T, error)
```

Subscribers first call `Unmarshal` to get the event type, then `UnmarshalData[TickData]` or `UnmarshalData[KlineData]` based on the type. This keeps `pkg/eventbus/` free of any domain types.

### Symbol Normalization

Exchanges use different symbol formats. Each adapter normalizes to the canonical format (uppercase, no separator):

| Exchange | Native Format | Canonical |
|----------|--------------|-----------|
| Binance | `BTCUSDT` | `BTCUSDT` (already canonical) |
| OKX | `BTC-USDT` | `BTCUSDT` |
| Bybit | `BTCUSDT` | `BTCUSDT` |
| Gate.io | `BTC_USDT` | `BTCUSDT` |

Normalization happens in each adapter's `Parse` method. Reverse mapping (canonical → exchange-native) is handled in `BuildSubscribeMsg`.

## EventBus (pkg/eventbus/)

### Interface

```go
type Bus interface {
    Publish(ctx context.Context, subject string, data []byte) error
    Subscribe(subject string, handler Handler) (Subscription, error)
    // Drain flushes pending messages to subscribers, then closes.
    // In-memory: delivers buffered messages, closes channels.
    // NATS: calls nc.Drain() to flush and close gracefully.
    Drain() error
    Close() error
}

type Handler func(subject string, data []byte) error

type Subscription interface {
    Unsubscribe() error
}
```

Abstracted as a reusable `pkg/eventbus/` package with no business logic. Two implementations selected via config:

### In-Memory Implementation

- `sync.RWMutex` + channel map
- Configurable buffer size (default 1024)
- Slow consumer strategy: drop newest event + log warning
- Simple wildcard matching: `*` single-level, `>` multi-level (matching NATS semantics)
- Use case: development, testing, single-node lightweight deployment

### NATS JetStream Implementation

- Persistent message streams with configurable retention
- Consumer ACK support for reliable delivery
- Native wildcard subscription
- Use case: production, requires message durability

### NATS Stream Configuration

| Stream | Subjects | Retention |
|--------|----------|-----------|
| `PRICE_TICK` | `price.tick.>` | 1 hour |
| `PRICE_KLINE` | `price.kline.>` | 24 hours |

### NATS Consumer Configuration

| Consumer | Stream | Deliver Policy | ACK |
|----------|--------|---------------|-----|
| CacheSubscriber | PRICE_TICK | DeliverLast | No (at-most-once) |
| StorageSubscriber | PRICE_TICK + PRICE_KLINE | DeliverAll | Manual (at-least-once) |

### Subject Design

```
price.tick.{exchange}.{symbol}     e.g. price.tick.binance.BTCUSDT
price.kline.{exchange}.{symbol}    e.g. price.kline.okx.ETHUSDT
```

Hierarchical subjects enable flexible filtering: `price.tick.>` for all ticks, `price.tick.binance.>` for Binance only.

## Exchange Adapter Layer

### Interface

```go
type ExchangeAdapter interface {
    Name() string
    Subscribe(symbols []string) error
    Unsubscribe(symbols []string) error
    FetchKlines(ctx context.Context, symbol, interval string, start, end time.Time) ([]KlineData, error)
    Start(ctx context.Context) error
    Stop() error
}
```

- `FetchKlines` returns `[]KlineData` (domain data), not event envelopes. The caller wraps in `Event` if needed for publishing or uses directly for API responses.
- `FetchKlines` accepts `context.Context` for timeout and cancellation propagation on REST calls.
- `Start(ctx)` is bridged to the `app.Server` interface via a wrapper struct that captures the context at construction.

### BaseAdapter (shared infrastructure)

Common WebSocket logic extracted into `BaseAdapter`. Uses callback functions for exchange-specific behavior:

```go
type MessageParser func(msg []byte) (any, error)              // parse exchange format → TickData/KlineData
type MessageBuilder func(symbols []string) ([]byte, error)    // build subscribe/unsubscribe message
type PingBuilder func() []byte                                // build ping payload

type BaseAdapter struct {
    name         string
    bus          eventbus.Bus
    logger       *slog.Logger
    parser       MessageParser
    subBuilder   MessageBuilder   // builds subscribe message
    unsubBuilder MessageBuilder   // builds unsubscribe message
    pingBuilder  PingBuilder
    cfg          ExchangeConfig
    // ... connection state
}
```

BaseAdapter provides:
- Connection management (connect, reconnect with exponential backoff)
- Heartbeat/ping-pong handling (interval from config)
- Event publishing to EventBus (using `Event.BuildSubject()`)
- Read loop with message dispatch to `parser`

Each exchange adapter constructs a `BaseAdapter` with its specific callbacks and config.

### Exchange-Specific Considerations

| Exchange | WebSocket Notes |
|----------|----------------|
| Binance | Multi-stream per connection (`<symbol>@trade`), 24h auto-disconnect requires reconnect |
| OKX | Subscribe to `tickers` channel, custom ping/pong interval |
| Bybit | `publicTrade.<symbol>` topic format |
| Gate.io | `spot.trades` channel, client-side ping frames required |

### REST API Rate Limiting

All four exchanges enforce rate limits on REST endpoints. Each adapter implements a per-exchange rate limiter (token bucket) to prevent IP bans:

| Exchange | Approximate Limit |
|----------|------------------|
| Binance | 1200 weight/min |
| OKX | 20 req/2s |
| Bybit | 120 req/min |
| Gate.io | 900 req/min |

Rate limiter configuration is per-adapter, using `golang.org/x/time/rate`.

## Subscriber Layer (Consumers)

### CacheSubscriber (Redis)

Subscribes to `price.tick.>`, writes latest prices to Redis.

**Data structures:**

| Key Pattern | Type | Content | TTL |
|-------------|------|---------|-----|
| `price:{exchange}:{symbol}` | Hash | price, volume, timestamp | 5 min |
| `price:all:{symbol}` | Sorted Set | member=exchange, score=price | 5 min |

- Hash: per-exchange per-symbol latest price for precise queries
- Sorted Set: cross-exchange price comparison for a symbol

Write strategy: direct overwrite on each event, no aggregation needed.

**Redis failure handling:** Log warning and continue. Cache is ephemeral — the next tick will overwrite. No retry, no circuit breaker. If Redis is down for an extended period, price queries return stale/empty data, which is acceptable since it is a cache layer.

### StorageSubscriber (InfluxDB)

Subscribes to `price.tick.>` + `price.kline.>`, batch writes to InfluxDB.

**InfluxDB data model:**

```
spot_tick,exchange=binance,symbol=BTCUSDT price=67000.50,volume=1.5 1711360000000000000
spot_kline,exchange=binance,symbol=BTCUSDT,interval=1m open=67000,high=67100,low=66950,close=67050,volume=120.5 1711360000000000000
```

- Tags (indexed): exchange, symbol, interval
- Fields (values): price, volume, open, high, low, close

**Batch write configuration:**

| Parameter | Value | Reasoning |
|-----------|-------|-----------|
| Batch size | 500 | Balance latency and IO efficiency |
| Flush interval | 5 seconds | Flush even if batch not full |
| Retry | Exponential backoff, max 3 | Handle transient InfluxDB unavailability |

**Retention (two buckets):**

| Bucket | Data | Retention |
|--------|------|-----------|
| `prices-tick` | spot_tick measurements | 7 days |
| `prices-kline` | spot_kline measurements | 1 year |

Two InfluxDB buckets are required because each bucket has a single retention policy.

## SubscriptionManager

Manages trading pair subscriptions, persisted to Redis, coordinates dynamic subscription with adapters.

```go
type SubscriptionManager interface {
    Add(ctx context.Context, req SubscribeRequest) (SubscribeResult, error)
    Remove(ctx context.Context, req UnsubscribeRequest) (SubscribeResult, error)
    List(ctx context.Context) ([]SubscriptionInfo, error)
    Restore(ctx context.Context) error
}

type SubscribeRequest struct {
    Exchanges []string  // empty means all enabled exchanges
    Symbols   []string
}

// SubscribeResult reports per-exchange outcome for partial failure handling
type SubscribeResult struct {
    Succeeded []string          // exchanges that subscribed successfully
    Failed    map[string]string // exchange → error message
}
```

**Redis storage:** `sub:{exchange}` (Set) — subscribed symbols per exchange.

**Flow:** API call → SubscriptionManager updates Redis → attempts subscription on each target adapter → returns `SubscribeResult` with per-exchange status.

**Partial failure semantics:** Redis is updated first (intent is persisted). Adapter subscription is best-effort per exchange. The API response includes per-exchange success/failure. On `Restore()`, all persisted subscriptions are retried, so transient adapter failures self-heal on restart.

**Concurrency:** `SubscriptionManager` uses an internal `sync.Mutex` to serialize Add/Remove operations, preventing race conditions from concurrent API calls.

**Restore:** On startup, reads all `sub:*` keys from Redis and notifies each adapter to re-subscribe.

## HTTP API (Gin)

### Subscription Management

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/subscriptions` | Add trading pair subscriptions |
| `DELETE` | `/api/v1/subscriptions` | Remove trading pair subscriptions |
| `GET` | `/api/v1/subscriptions` | List all subscriptions |

### Price Queries

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/prices/:symbol` | Latest price across exchanges (Redis) |
| `GET` | `/api/v1/prices` | All subscribed symbols latest prices |

### K-line Queries

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/klines/:symbol` | Historical K-lines (InfluxDB) |

Query params: `exchange`, `interval` (default 1m), `start`, `end`, `limit` (default 500)

K-lines are fetched on-demand: when the API is called, the handler queries InfluxDB for stored data. If data is not available in InfluxDB (e.g., first request for a new pair), the handler falls back to calling the adapter's `FetchKlines` REST method, stores the result, and returns it.

### Health Check

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Liveness check (always 200) |
| `GET` | `/ready` | Readiness check (Redis + InfluxDB + at least one adapter connected) |

### Response Format

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

### HTTP Server Lifecycle

A `GinServer` struct wraps `*gin.Engine` and `*http.Server`, implementing `app.Server`:

```go
type GinServer struct {
    engine *gin.Engine
    server *http.Server
}

func (s *GinServer) Start() error  { return s.server.ListenAndServe() }
func (s *GinServer) Stop() error {
    ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
    defer cancel()
    return s.server.Shutdown(ctx)
}
func (s *GinServer) String() string { return "http-server" }
```

Registered in `CreateServices` alongside exchange adapter server wrappers.

## ServiceContext Changes

```go
type ServiceContext struct {
    Config   config.Config
    Logger   *slog.Logger
    Redis    *redisx.Client
    Influx   *influxdb.Client
    Bus      eventbus.Bus
    SubMgr   SubscriptionManager
    Adapters []ExchangeAdapter
}
```

MySQL removed from ServiceContext and Config. The `pkg/database/mysqlx/` package is preserved in the codebase for potential future use but not referenced by this service.

## Configuration (watch.yaml)

```yaml
log:
  level: "info"
  format: "json"
  output: "stdout"
  addSource: false
  enableTrace: true
  timeFormat: "2006-01-02 15:04:05"

redis:
  addrs: ["127.0.0.1:6379"]
  password: ""
  db: 0

influxdb:
  url: "http://127.0.0.1:8086"
  token: ""
  org: "price-watch"
  buckets:
    tick: "prices-tick"
    kline: "prices-kline"
  batchSize: 500
  flushInterval: "5s"

eventbus:
  driver: "memory"
  memory:
    bufferSize: 1024
  nats:
    url: "nats://127.0.0.1:4222"
    streams:
      - name: "PRICE_TICK"
        subjects: ["price.tick.>"]
        maxAge: "1h"
      - name: "PRICE_KLINE"
        subjects: ["price.kline.>"]
        maxAge: "24h"

http:
  addr: ":8080"

exchanges:
  binance:
    enabled: true
    wsUrl: "wss://stream.binance.com:9443/ws"
    restUrl: "https://api.binance.com"
    reconnectBase: "1s"
    reconnectMax: "60s"
    pingInterval: "30s"
  okx:
    enabled: true
    wsUrl: "wss://ws.okx.com:8443/ws/v5/public"
    restUrl: "https://www.okx.com"
    reconnectBase: "1s"
    reconnectMax: "60s"
    pingInterval: "25s"
  bybit:
    enabled: true
    wsUrl: "wss://stream.bybit.com/v5/public/spot"
    restUrl: "https://api.bybit.com"
    reconnectBase: "1s"
    reconnectMax: "60s"
    pingInterval: "20s"
  gateio:
    enabled: true
    wsUrl: "wss://api.gateio.ws/ws/v4/"
    restUrl: "https://api.gateio.ws"
    reconnectBase: "1s"
    reconnectMax: "60s"
    pingInterval: "15s"
```

### Exchange Config Go Struct

```go
type ExchangeConfig struct {
    Enabled       bool          `mapstructure:"enabled"`
    WsURL         string        `mapstructure:"wsUrl"`
    RestURL       string        `mapstructure:"restUrl"`
    ReconnectBase time.Duration `mapstructure:"reconnectBase"`
    ReconnectMax  time.Duration `mapstructure:"reconnectMax"`
    PingInterval  time.Duration `mapstructure:"pingInterval"`
}
```

## Project Directory Structure (new additions)

```
pkg/
├── eventbus/                   # Generic EventBus abstraction
│   ├── eventbus.go             # Bus interface + factory
│   ├── config.go
│   ├── memory.go
│   └── nats.go
└── database/
    ├── redisx/                 # Existing
    ├── mysqlx/                 # Existing (preserved, not used by this service)
    └── influxdb/               # New: InfluxDB client wrapper
        ├── config.go
        └── influxdb.go

internal/watch/
├── config/config.go            # Extended: influxdb, eventbus, http, exchanges sections
├── svc/serviceContext.go       # Reworked: remove MySQL, add Influx/Bus/SubMgr/Adapters
├── event/                      # New: event model + business serialization
│   ├── event.go                # Event[T], TickData, KlineData, BuildSubject()
│   ├── marshal.go              # Marshal/Unmarshal helpers (JSON, two-pass decode)
│   ├── publisher.go
│   └── subscriber.go
├── exchange/                   # New: exchange adapters
│   ├── adapter.go              # ExchangeAdapter interface + BaseAdapter + callbacks
│   ├── binance.go
│   ├── okx.go
│   ├── bybit.go
│   └── gateio.go
├── subscription/               # New: subscription management
│   └── manager.go
├── consumer/                   # New: EventBus consumers
│   ├── cache.go                # CacheSubscriber → Redis
│   └── storage.go              # StorageSubscriber → InfluxDB
└── handler/                    # New: HTTP handlers
    ├── subscription.go
    ├── price.go
    └── kline.go

cmd/watch/
├── main.go
└── initial/
    ├── initApp.go              # Start consumers, restore subscriptions
    ├── createService.go        # Register GinServer + Adapter server wrappers
    └── close.go                # Ordered cleanup registration
```

## Docker Infrastructure Additions

Add to `docker-compose.yml`:

```yaml
influxdb:
  image: influxdb:2.7-alpine
  ports:
    - "8086:8086"
  volumes:
    - ${DATA_DIR}/influxdb:/var/lib/influxdb2

nats:
  image: nats:2.10-alpine
  command: ["--jetstream", "--store_dir", "/data"]
  ports:
    - "4222:4222"
    - "8222:8222"
  volumes:
    - ${DATA_DIR}/nats:/data
```

MySQL service is preserved in docker-compose.yml but not required for this service.

## Future Iterations

- Alert notifications (WebSocket push, Webhook, Telegram Bot) — subscribe to EventBus
- Futures/derivatives price support
- More exchanges
- K-line aggregation from tick data
