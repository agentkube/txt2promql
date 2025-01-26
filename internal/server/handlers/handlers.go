package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/agentkube/txt2promql/internal/agent"
	"github.com/agentkube/txt2promql/internal/prometheus"
	"github.com/agentkube/txt2promql/internal/provider/openai"
	"github.com/labstack/echo/v4"
)

type Handlers struct {
	promClient       *prometheus.Client
	contextExtractor *agent.ContextExtractor
	explainer        *agent.Explainer
	queryBuilder     *agent.QueryBuilder
	metricCache      map[string]prometheus.MetricSchema
	lastCacheTime    time.Time
}

func New(promClient *prometheus.Client, openaiClient *openai.OpenAIClient) *Handlers {
	return &Handlers{
		promClient:       promClient,
		contextExtractor: agent.NewContextExtractor(openaiClient),
		explainer:        agent.NewExplainer(openaiClient),
		queryBuilder:     agent.NewQueryBuilder(),
		metricCache:      make(map[string]prometheus.MetricSchema),
	}
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
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	if strings.TrimSpace(req.Query) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Query cannot be empty")
	}

	if err := h.refreshMetricCache(c.Request().Context()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to refresh metrics")
	}

	queryCtx, err := h.contextExtractor.ExtractQueryContext(c.Request().Context(), req.Query, h.metricCache)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process query")
	}

	promQL, warnings := h.queryBuilder.Build(queryCtx)
	if promQL == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Could not generate valid PromQL")
	}

	validationResult, err := h.promClient.ValidateQuery(c.Request().Context(), promQL)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Query validation error: %v", err))
	}
	if !validationResult.Valid {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid PromQL query: %s", validationResult.Error))
	}

	explanation := h.explainer.GenerateExplanation(c.Request().Context(), queryCtx, promQL)

	resp := ConvertResponse{
		PromQL:      promQL,
		Explanation: explanation,
		Warnings:    warnings,
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *Handlers) refreshMetricCache(ctx context.Context) error {
	if time.Since(h.lastCacheTime) < 5*time.Minute {
		return nil
	}

	result, err := h.promClient.Query(ctx, "{__name__=~\".+\"}")
	if err != nil {
		return err
	}

	newCache := make(map[string]prometheus.MetricSchema)
	for _, m := range result.Data.Result {
		if name, ok := m.Metric["__name__"]; ok {
			newCache[name] = prometheus.MetricSchema{
				Name:   name,
				Labels: m.Metric,
			}
		}
	}

	h.metricCache = newCache
	h.lastCacheTime = time.Now()
	return nil
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
