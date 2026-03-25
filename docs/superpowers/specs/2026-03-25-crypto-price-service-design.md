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
| Historical storage | InfluxDB | Purpose-built for time-series data |
| Event distribution | EventBus (pluggable: in-memory / NATS JetStream) | Decoupled consumers, configurable reliability |
| HTTP framework | Gin | Mature ecosystem, middleware-rich, high performance |
| Subscription management | API-driven, persisted to Redis | Runtime flexibility without restart |
| K-line intervals | 1m, 5m, 15m, 1h, 4h, 1d | Standard intervals supported by all four exchanges |
| Alert notifications | Deferred to future iteration | Focus on data acquisition and storage first |
| MySQL | Not used in this phase | No relational data needs yet |

## Architecture

### Event-Driven with EventBus

```
Exchange Adapters ──→ EventBus ──→ CacheSubscriber (Redis)
                             ──→ StorageSubscriber (InfluxDB)
                             ──→ (future: AlertSubscriber)
```

All exchange adapters publish unified events to the EventBus. Subscribers consume events independently. Adding new consumers (e.g., alerts) requires only subscribing to the bus — no changes to adapters or existing consumers.

## Event Model

### Unified Envelope (CloudEvents-inspired, Go generics)

```go
type Event[T any] struct {
    ID        string    `json:"id"`
    Type      string    `json:"type"`      // "price.tick", "price.kline"
    Source    string    `json:"source"`    // "binance", "okx", "bybit", "gateio"
    Subject   string    `json:"subject"`   // "BTCUSDT"
    Timestamp time.Time `json:"timestamp"`
    Data      T         `json:"data"`
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
- Serialization/deserialization handled at business layer, not in EventBus

## EventBus (pkg/eventbus/)

### Interface

```go
type Bus interface {
    Publish(ctx context.Context, subject string, data []byte) error
    Subscribe(subject string, handler Handler) (Subscription, error)
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
    FetchKlines(symbol, interval string, start, end time.Time) ([]Event[KlineData], error)
    Start(ctx context.Context) error
    Stop() error
}
```

### BaseAdapter (shared infrastructure)

Common WebSocket logic extracted into `BaseAdapter`:

- Connection management (connect, reconnect with exponential backoff)
- Heartbeat/ping-pong handling
- Event publishing to EventBus
- Read loop with message dispatch

Each exchange adapter composes `BaseAdapter` and implements:

- `Parse(msg []byte) (Event, error)` — parse exchange-specific format
- `BuildSubscribeMsg(symbols) []byte` — construct subscribe message
- `BuildUnsubscribeMsg(symbols) []byte` — construct unsubscribe message
- `PingPayload() []byte` — exchange-specific heartbeat content

### Exchange-Specific Considerations

| Exchange | WebSocket Notes |
|----------|----------------|
| Binance | Multi-stream per connection (`<symbol>@trade`), 24h auto-disconnect requires reconnect |
| OKX | Subscribe to `tickers` channel, custom ping/pong interval |
| Bybit | `publicTrade.<symbol>` topic format |
| Gate.io | `spot.trades` channel, client-side ping frames required |

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

**Retention:** tick data 7 days, kline data 1 year (configured via InfluxDB buckets).

## SubscriptionManager

Manages trading pair subscriptions, persisted to Redis, coordinates dynamic subscription with adapters.

```go
type SubscriptionManager interface {
    Add(ctx context.Context, req SubscribeRequest) error
    Remove(ctx context.Context, req UnsubscribeRequest) error
    List(ctx context.Context) ([]SubscriptionInfo, error)
    Restore(ctx context.Context) error
}

type SubscribeRequest struct {
    Exchanges []string  // empty means all exchanges
    Symbols   []string
}
```

**Redis storage:** `sub:{exchange}` (Set) — subscribed symbols per exchange.

**Flow:** API call → SubscriptionManager updates Redis → notifies target Adapter to Subscribe/Unsubscribe.

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

### Response Format

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

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

MySQL removed. InfluxDB, EventBus, SubscriptionManager, and Adapters added.

## Configuration (watch.yaml)

```yaml
log:
  level: "info"
  format: "json"
  output: "stdout"
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
  bucket: "prices"
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
  okx:
    enabled: true
    wsUrl: "wss://ws.okx.com:8443/ws/v5/public"
    restUrl: "https://www.okx.com"
  bybit:
    enabled: true
    wsUrl: "wss://stream.bybit.com/v5/public/spot"
    restUrl: "https://api.bybit.com"
  gateio:
    enabled: true
    wsUrl: "wss://api.gateio.ws/ws/v4/"
    restUrl: "https://api.gateio.ws"
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
    └── influxdb/               # New: InfluxDB client wrapper
        ├── config.go
        └── influxdb.go

internal/watch/
├── config/config.go            # Extended: influxdb, eventbus, http, exchanges sections
├── svc/serviceContext.go       # Reworked: remove MySQL, add Influx/Bus/SubMgr/Adapters
├── event/                      # New: event model + business serialization
│   ├── event.go                # Event[T], TickData, KlineData
│   ├── publisher.go
│   └── subscriber.go
├── exchange/                   # New: exchange adapters
│   ├── adapter.go              # ExchangeAdapter interface + BaseAdapter
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

## Future Iterations

- Alert notifications (WebSocket push, Webhook, Telegram Bot) — subscribe to EventBus
- Futures/derivatives price support
- More exchanges
- K-line aggregation from tick data
