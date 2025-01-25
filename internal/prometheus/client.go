package prometheus

import (
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

type Client struct {
	api v1.API
}

func NewClient(url string) *Client {
	cfg := api.Config{Address: url}
	client, _ := api.NewClient(cfg)
	return &Client{api: v1.NewAPI(client)}
}

func (c *Client) GetMetrics() []string {
	// Fetch metrics via /api/v1/labels

	return []string{}
}
