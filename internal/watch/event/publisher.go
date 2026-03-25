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
