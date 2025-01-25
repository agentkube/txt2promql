// internal/server/server.go
package server

import (
	"github.com/agentkube/txt2promql/internal/prometheus"
	"github.com/agentkube/txt2promql/internal/server/handlers"
	"github.com/labstack/echo/v4"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	requestDuration = promauto.NewHistogramVec(
		prom.HistogramOpts{
			Name:    "text2promql_request_duration_seconds",
			Help:    "Time spent processing request",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"handler", "status"},
	)

	requestTotal = promauto.NewCounterVec(
		prom.CounterOpts{
			Name: "text2promql_requests_total",
			Help: "Total number of requests",
		},
		[]string{"handler", "status"},
	)
)

func RegisterHandlers(e *echo.Echo, promClient *prometheus.Client) {
	h := handlers.New(promClient)
	e.POST("/api/v1/convert", h.HandleConvert)
	e.POST("/api/v1/validate", h.HandleValidate)
	e.GET("/api/v1/metrics", h.HandleListMetrics)
	e.GET("/health", h.HandleHealth)
}
