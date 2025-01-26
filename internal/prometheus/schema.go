// internal/prometheus/schema.go
package prometheus

import (
	"context"
	"fmt"
	"time"
)

type MetricSchema struct {
	Name       string            `json:"name"`
	Type       string            `json:"type"` // counter, gauge, histogram
	Help       string            `json:"help"`
	Labels     map[string]string `json:"labels"`
	LastScrape time.Time         `json:"last_scrape"`
}

type SchemaManager struct {
	client *Client
	cache  map[string]MetricSchema
}

func NewSchemaManager(client *Client) *SchemaManager {
	return &SchemaManager{
		client: client,
		cache:  make(map[string]MetricSchema),
	}
}

func (sm *SchemaManager) RefreshMetrics(ctx context.Context) error {
	result, err := sm.client.Query(ctx, "{__name__=~\".+\"}")
	if err != nil {
		return fmt.Errorf("querying metrics: %w", err)
	}

	for _, metric := range result.Data.Result {
		name := metric.Metric["__name__"]
		sm.cache[name] = MetricSchema{
			Name:       name,
			Labels:     metric.Metric,
			LastScrape: time.Now(),
		}
	}

	return nil
}
