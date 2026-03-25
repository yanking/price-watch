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
