package consumer

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	influxdb2write "github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/yanking/price-watch/internal/watch/event"
	"github.com/yanking/price-watch/pkg/database/influxdb"
	"github.com/yanking/price-watch/pkg/eventbus"
)

type StorageSubscriber struct {
	influx        *influxdb.Client
	bus           eventbus.Bus
	logger        *slog.Logger
	batchSize     int
	flushInterval time.Duration

	mu       sync.Mutex
	tickBuf  []*influxdb2write.Point
	klineBuf []*influxdb2write.Point
	tickSub  eventbus.Subscription
	klineSub eventbus.Subscription
	stopCh   chan struct{}
	stopped  chan struct{}
}

func NewStorageSubscriber(
	influx *influxdb.Client,
	bus eventbus.Bus,
	logger *slog.Logger,
	batchSize int,
	flushInterval time.Duration,
) *StorageSubscriber {
	if batchSize <= 0 {
		batchSize = 500
	}
	if flushInterval <= 0 {
		flushInterval = 5 * time.Second
	}
	return &StorageSubscriber{
		influx:        influx,
		bus:           bus,
		logger:        logger.With("component", "storage-subscriber"),
		batchSize:     batchSize,
		flushInterval: flushInterval,
		stopCh:        make(chan struct{}),
		stopped:       make(chan struct{}),
	}
}

func (s *StorageSubscriber) Start() error {
	tickSub, err := event.SubscribeTick(s.bus, "price.tick.>", s.handleTick)
	if err != nil {
		return fmt.Errorf("subscribe tick: %w", err)
	}
	s.tickSub = tickSub

	klineSub, err := event.SubscribeKline(s.bus, "price.kline.>", s.handleKline)
	if err != nil {
		return fmt.Errorf("subscribe kline: %w", err)
	}
	s.klineSub = klineSub

	go s.flushLoop()
	s.logger.Info("storage subscriber started")
	return nil
}

func (s *StorageSubscriber) Stop() error {
	close(s.stopCh)
	<-s.stopped
	if s.tickSub != nil {
		s.tickSub.Unsubscribe()
	}
	if s.klineSub != nil {
		s.klineSub.Unsubscribe()
	}
	// Final flush
	s.flush()
	return nil
}

func (s *StorageSubscriber) handleTick(e event.Event[event.TickData]) error {
	priceFloat, _ := e.Data.Price.Float64()
	volFloat, _ := e.Data.Volume.Float64()

	point := influxdb2write.NewPoint(
		"spot_tick",
		map[string]string{
			"exchange": e.Source,
			"symbol":   e.Subject,
		},
		map[string]interface{}{
			"price":  priceFloat,
			"volume": volFloat,
		},
		e.Timestamp,
	)

	s.mu.Lock()
	s.tickBuf = append(s.tickBuf, point)
	shouldFlush := len(s.tickBuf) >= s.batchSize
	s.mu.Unlock()

	if shouldFlush {
		s.flushTick()
	}
	return nil
}

func (s *StorageSubscriber) handleKline(e event.Event[event.KlineData]) error {
	openF, _ := e.Data.Open.Float64()
	highF, _ := e.Data.High.Float64()
	lowF, _ := e.Data.Low.Float64()
	closeF, _ := e.Data.Close.Float64()
	volF, _ := e.Data.Volume.Float64()

	point := influxdb2write.NewPoint(
		"spot_kline",
		map[string]string{
			"exchange": e.Source,
			"symbol":   e.Subject,
			"interval": e.Data.Interval,
		},
		map[string]interface{}{
			"open":   openF,
			"high":   highF,
			"low":    lowF,
			"close":  closeF,
			"volume": volF,
		},
		e.Timestamp,
	)

	s.mu.Lock()
	s.klineBuf = append(s.klineBuf, point)
	shouldFlush := len(s.klineBuf) >= s.batchSize
	s.mu.Unlock()

	if shouldFlush {
		s.flushKline()
	}
	return nil
}

func (s *StorageSubscriber) flushLoop() {
	defer close(s.stopped)
	ticker := time.NewTicker(s.flushInterval)
	defer ticker.Stop()
	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.flush()
		}
	}
}

func (s *StorageSubscriber) flush() {
	s.flushTick()
	s.flushKline()
}

func (s *StorageSubscriber) flushTick() {
	s.mu.Lock()
	if len(s.tickBuf) == 0 {
		s.mu.Unlock()
		return
	}
	points := s.tickBuf
	s.tickBuf = nil
	s.mu.Unlock()

	if s.influx == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.influx.WriteTickPoints(ctx, points...); err != nil {
		s.logger.Warn("flush tick points", "count", len(points), "error", err)
	}
}

func (s *StorageSubscriber) flushKline() {
	s.mu.Lock()
	if len(s.klineBuf) == 0 {
		s.mu.Unlock()
		return
	}
	points := s.klineBuf
	s.klineBuf = nil
	s.mu.Unlock()

	if s.influx == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.influx.WriteKlinePoints(ctx, points...); err != nil {
		s.logger.Warn("flush kline points", "count", len(points), "error", err)
	}
}
