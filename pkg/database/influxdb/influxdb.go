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
