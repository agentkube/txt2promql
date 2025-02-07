// internal/core/knowledgegraph/client.go
package knowledgegraph

import (
	"context"
	"fmt"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Client struct {
	enabled bool
	driver  neo4j.DriverWithContext
	config  *Config
}

type Config struct {
	Enabled  bool
	URI      string
	User     string
	Password string
	Database string
}

type MetricInfo struct {
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Description string            `json:"description"`
	Labels      map[string]string `json:"labels"`
	SimilarTo   []string          `json:"similar_to,omitempty"`
}

// NewClient creates a new knowledge graph client
func NewClient(config *Config) (*Client, error) {
	client := &Client{
		enabled: config.Enabled,
		config:  config,
	}

	if !config.Enabled {
		return client, nil
	}

	driver, err := neo4j.NewDriverWithContext(
		config.URI,
		neo4j.BasicAuth(config.User, config.Password, ""),
		func(c *neo4j.Config) {
			c.MaxConnectionLifetime = 30 * time.Minute
			c.MaxConnectionPoolSize = 50
		})
	if err != nil {
		return nil, fmt.Errorf("creating neo4j driver: %w", err)
	}

	client.driver = driver
	return client, nil
}

func (c *Client) Close(ctx context.Context) error {
	if !c.enabled || c.driver == nil {
		return nil
	}
	return c.driver.Close(ctx)
}

func (c *Client) Connect(ctx context.Context) error {
	if !c.enabled || c.driver == nil {
		return nil
	}

	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		err := c.driver.VerifyConnectivity(ctx)
		if err == nil {
			return nil
		}
		fmt.Printf("Attempt %d: Failed to connect to Neo4j: %v\n", i+1, err)
		time.Sleep(time.Second * 2) // Wait before retrying
	}

	return c.driver.VerifyConnectivity(ctx)
}

func (c *Client) FindSimilarMetrics(ctx context.Context, metricName string) ([]MetricInfo, error) {
	if !c.enabled || c.driver == nil {
		return nil, nil
	}

	session := c.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: c.config.Database,
		AccessMode:   neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := neo4j.ExecuteRead[[]MetricInfo](ctx, session,
		func(tx neo4j.ManagedTransaction) ([]MetricInfo, error) {
			query := `
            MATCH (m:Metric {name: $name})-[r:RELATED]-(similar:Metric)
            RETURN similar, r.weight as weight
            ORDER BY r.weight DESC
            LIMIT 5
            `
			result, err := tx.Run(ctx, query, map[string]any{"name": metricName})
			if err != nil {
				return nil, err
			}

			var metrics []MetricInfo
			records, err := result.Collect(ctx)
			if err != nil {
				return nil, err
			}

			for _, record := range records {
				node, ok := record.Values[0].(neo4j.Node)
				if !ok {
					continue
				}

				metric := MetricInfo{
					Name:        node.Props["name"].(string),
					Type:        node.Props["type"].(string),
					Description: node.Props["description"].(string),
				}
				if labels, ok := node.Props["labels"].(map[string]string); ok {
					metric.Labels = labels
				}
				metrics = append(metrics, metric)
			}

			return metrics, nil
		})

	if err != nil {
		return nil, fmt.Errorf("finding similar metrics: %w", err)
	}

	return result, nil
}
