// internal/server/handlers/handlers.go
package handlers

import (
	"context"
	"net/http"

	"github.com/agentkube/txt2promql/internal/prometheus"
	"github.com/labstack/echo/v4"
)

type Handlers struct {
	promClient *prometheus.Client
}

func New(promClient *prometheus.Client) *Handlers {
	return &Handlers{promClient: promClient}
}

type ConvertRequest struct {
	Query string `json:"query"`
}

type ConvertResponse struct {
	PromQL      string   `json:"promql"`
	Explanation string   `json:"explanation,omitempty"`
	Warnings    []string `json:"warnings,omitempty"`
}

func (h *Handlers) HandleConvert(c echo.Context) error {
	var req ConvertRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	resp := ConvertResponse{
		PromQL: "rate(http_requests_total{status=~\"5..\"}[5m])",
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *Handlers) HandleValidate(c echo.Context) error {
	var req struct {
		PromQL string `json:"promql"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	result, err := h.promClient.ValidateQuery(c.Request().Context(), req.PromQL)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

func (h *Handlers) HandleListMetrics(c echo.Context) error {
	result, err := h.promClient.Query(c.Request().Context(), "{__name__=~\".+\"}")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	metrics := make([]string, 0)
	for _, m := range result.Data.Result {
		if name, ok := m.Metric["__name__"]; ok {
			metrics = append(metrics, name)
		}
	}

	return c.JSON(http.StatusOK, metrics)
}

func (h *Handlers) HandleHealth(c echo.Context) error {
	ctx := context.Background()

	_, err := h.promClient.Query(ctx, "up")
	promOK := err == nil

	resp := map[string]interface{}{
		"status":     "ok",
		"prometheus": promOK,
	}
	if !promOK {
		resp["status"] = "degraded"
	}

	return c.JSON(http.StatusOK, resp)
}
