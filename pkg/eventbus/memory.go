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
