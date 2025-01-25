package prometheus

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Discovery struct {
	client      *Client
	schemas     map[string]MetricSchema
	mu          sync.RWMutex
	lastRefresh time.Time
}

func NewDiscovery(client *Client) *Discovery {
	return &Discovery{
		client:  client,
		schemas: make(map[string]MetricSchema),
	}
}

func (d *Discovery) RefreshMetrics(ctx context.Context) error {
	result, err := d.client.Query(ctx, "{__name__=~\".+\"}")
	if err != nil {
		return fmt.Errorf("querying metrics: %w", err)
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	for _, metric := range result.Data.Result {
		name := metric.Metric["__name__"]
		d.schemas[name] = MetricSchema{
			Name:       name,
			Labels:     metric.Metric,
			LastScrape: time.Now(),
		}
	}

	d.lastRefresh = time.Now()
	return nil
}

func (d *Discovery) GetMetric(name string) (MetricSchema, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	schema, ok := d.schemas[name]
	return schema, ok
}
