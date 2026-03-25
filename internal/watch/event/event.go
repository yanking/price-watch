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
