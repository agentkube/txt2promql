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

type ExecuteRequest struct {
	Query     string     `json:"query"`
	Start     *time.Time `json:"start,omitempty"`
	End       *time.Time `json:"end,omitempty"`
	Step      string     `json:"step,omitempty"`
	Timestamp *time.Time `json:"timestamp,omitempty"`
}

// func (h *Handlers) HandleExecute(c echo.Context) error {
// 	var req ExecuteRequest
// 	if err := c.Bind(&req); err != nil {
// 		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
// 	}

// 	if strings.TrimSpace(req.Query) == "" {
// 		return echo.NewHTTPError(http.StatusBadRequest, "Query cannot be empty")
// 	}

// 	ctx := c.Request().Context()

// 	// If start and end times are provided, perform a range query
// 	if req.Start != nil && req.End != nil {
// 		step := time.Minute // default step
// 		if req.Step != "" {
// 			var err error
// 			step, err = time.ParseDuration(req.Step)
// 			if err != nil {
// 				return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid step: %v", err))
// 			}
// 		}

// 		result, err := h.promClient.QueryRange(ctx, req.Query, *req.Start, *req.End, step)
// 		if err != nil {
// 			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Range query error: %v", err))
// 		}
// 		return c.JSON(http.StatusOK, result)
// 	}

// 	// Otherwise, perform an instant query
// 	result, err := h.promClient.QueryInstant(ctx, req.Query, req.Timestamp)
// 	if err != nil {
// 		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Instant query error: %v", err))
// 	}

// 	return c.JSON(http.StatusOK, result)
// }

type ChartSuggestion struct {
	Type   string `json:"type"`
	Reason string `json:"reason"`
}

type ExecuteResponse struct {
	Status         string                  `json:"status"`
	Data           *prometheus.QueryResult `json:"data"`
	SuggestedChart ChartSuggestion         `json:"suggestedChart"`
}

func predictChartType(query string, result *prometheus.QueryResult) ChartSuggestion {
	// Check for time-based functions
	if strings.Contains(query, "rate(") ||
		strings.Contains(query, "increase(") ||
		strings.Contains(query, "irate(") {
		return ChartSuggestion{
			Type:   "time-series",
			Reason: "Query contains rate or increase functions which are best visualized over time",
		}
	}

	// Check for aggregations
	if strings.Contains(query, " by (") {
		if strings.Contains(query, "topk") || strings.Contains(query, "bottomk") {
			return ChartSuggestion{
				Type:   "bar",
				Reason: "Query uses top/bottom k aggregation which is best shown as a bar chart",
			}
		}
		return ChartSuggestion{
			Type:   "pie",
			Reason: "Query uses grouping which can be effectively shown as a pie chart",
		}
	}

	// Check for single value or gauge-like metrics
	if result.Data.ResultType == "vector" && len(result.Data.Result) == 1 {
		if strings.Contains(strings.ToLower(query), "total") ||
			strings.Contains(strings.ToLower(query), "sum") {
			return ChartSuggestion{
				Type:   "gauge",
				Reason: "Query returns a single total/sum value suitable for a gauge",
			}
		}
	}

	// Check for complex multi-metric results
	if result.Data.ResultType == "matrix" ||
		(result.Data.ResultType == "vector" && len(result.Data.Result) > 10) {
		return ChartSuggestion{
			Type:   "table",
			Reason: "Query returns multiple metrics which are best viewed in a table",
		}
	}

	// Check for hierarchical data
	if strings.Contains(query, "count") && strings.Contains(query, " by (") {
		return ChartSuggestion{
			Type:   "tree",
			Reason: "Query counts by labels which can be shown hierarchically",
		}
	}

	// Default to time-series for most other cases
	return ChartSuggestion{
		Type:   "time-series",
		Reason: "Default visualization for metric data over time",
	}
}

func (h *Handlers) HandleExecute(c echo.Context) error {
	var req ExecuteRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	if strings.TrimSpace(req.Query) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Query cannot be empty")
	}

	ctx := c.Request().Context()
	var result *prometheus.QueryResult
	var err error

	if req.Start != nil && req.End != nil {
		step := time.Minute
		if req.Step != "" {
			step, err = time.ParseDuration(req.Step)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid step: %v", err))
			}
		}
		result, err = h.promClient.QueryRange(ctx, req.Query, *req.Start, *req.End, step)
	} else {
		result, err = h.promClient.QueryInstant(ctx, req.Query, req.Timestamp)
	}

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Predict chart type based on query and result
	chartSuggestion := predictChartType(req.Query, result)

	response := ExecuteResponse{
		Status:         "success",
		Data:           result,
		SuggestedChart: chartSuggestion,
	}

	return c.JSON(http.StatusOK, response)
}
