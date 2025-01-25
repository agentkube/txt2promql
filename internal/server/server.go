// internal/server/server.go
package server

import (
	"net/http"
	"time"

	prometheus "github.com/agentkube/txt2promql/internal/prometheus"
	handlers "github.com/agentkube/txt2promql/internal/server/handlers"
	"github.com/labstack/echo/v4"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

// MetricsMiddleware records request metrics
func MetricsMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()

		err := next(c)

		// Record metrics
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

func RegisterHandlers(e *echo.Echo, promClient *prometheus.Client) {
	//  handlers
	h := handlers.New(promClient)
	// middlewares
	e.Use(MetricsMiddleware)

	// Metrics
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
	// Health check
	e.GET("/health", h.HandleHealth)

	api := e.Group("/api/v1")
	{
		api.POST("/convert", h.HandleConvert)
		api.POST("/validate", h.HandleValidate)
		api.GET("/metrics", h.HandleListMetrics)
	}
}
