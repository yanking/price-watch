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
