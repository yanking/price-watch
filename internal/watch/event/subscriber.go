package event

import (
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
		kline, err := UnmarshalData[KlineData](raw.Data)
		if err != nil {
			return err
		}
		return h(Event[KlineData]{
			ID: raw.ID, Type: raw.Type, Source: raw.Source,
			Subject: raw.Subject, Timestamp: raw.Timestamp, Data: kline,
		})
	})
}
