// internal/server/server.go
package server

import (
	"fmt"
	"net/http"
	"time"

	prometheus "github.com/agentkube/txt2promql/internal/prometheus"
	"github.com/agentkube/txt2promql/internal/provider/openai"
	handlers "github.com/agentkube/txt2promql/internal/server/handlers"
	"github.com/labstack/echo/v4"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
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

	errorTotal = promauto.NewCounterVec(
		prom.CounterOpts{
			Name: "text2promql_errors_total",
			Help: "Total number of errors",
		},
		[]string{"handler", "error_type"},
	)
)

func MetricsMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()
		err := next(c)

		duration := time.Since(start).Seconds()
		handler := c.Path()
		status := "success"

		if err != nil {
			status = "error"
			if he, ok := err.(*echo.HTTPError); ok {
				errorTotal.WithLabelValues(handler, http.StatusText(he.Code)).Inc()
			} else {
				errorTotal.WithLabelValues(handler, "internal").Inc()
			}
		}

		requestDuration.WithLabelValues(handler, status).Observe(duration)
		requestTotal.WithLabelValues(handler, status).Inc()

		return err
	}
}

func RegisterHandlers(e *echo.Echo, promClient *prometheus.Client) error {
	// Load AI configuration
	var aiConfig openai.Config
	if err := viper.UnmarshalKey("ai", &aiConfig); err != nil {
		return fmt.Errorf("loading AI configuration: %w", err)
	}

	// Initialize OpenAI client
	openaiClient, err := openai.NewClient(&aiConfig)
	if err != nil {
		return fmt.Errorf("initializing OpenAI client: %w", err)
	}

	h := handlers.New(promClient, openaiClient)
	// middleware
	e.Use(MetricsMiddleware)

	// routes
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
	e.GET("/health", h.HandleHealth)

	// API routes
	api := e.Group("/api/v1")
	{
		api.POST("/convert", h.HandleConvert) //TODO high chances of request failure when same statement flows in.
		api.POST("/validate", h.HandleValidate)
		api.POST("/execute", h.HandleExecute)
		api.GET("/metrics", h.HandleListMetrics)
	}

	return nil
}

//TODO high chances of request failure when same statement flows in.
// {
//   "message": "Failed to process query"
// }
// {
//   "message": "Invalid PromQL query: invalid query: unexpected status code: 400"
// }

// Agent Response:
// ```json
// [
//   {
//     "metric": "prometheus_http_request_duration_seconds_sum",
//     "labels": {"handler": "/api/v1/query_range"},
//     "timeRange": "5m",
//     "aggregation": "avg"
//   },
//   {
//     "metric": "prometheus_http_request_duration_seconds_count",
//     "labels": {"handler": "/api/v1/query_range"},
//     "timeRange": "5m",
//     "aggregation": "avg"
//   }
// ]
// ```
