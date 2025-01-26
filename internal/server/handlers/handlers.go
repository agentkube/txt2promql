package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/agentkube/txt2promql/internal/prometheus"
	"github.com/agentkube/txt2promql/internal/provider/openai"
	"github.com/agentkube/txt2promql/internal/types"
	"github.com/labstack/echo/v4"
)

type Handlers struct {
	promClient    *prometheus.Client
	openaiClient  *openai.OpenAIClient
	openaiProc    *openai.Processor
	metricCache   map[string]prometheus.MetricSchema
	lastCacheTime time.Time
}

func New(promClient *prometheus.Client, openaiClient *openai.OpenAIClient) *Handlers {
	return &Handlers{
		promClient:   promClient,
		openaiClient: openaiClient,
		openaiProc:   openai.NewProcessor(openaiClient),
		metricCache:  make(map[string]prometheus.MetricSchema),
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

	// Refresh metric cache if needed
	if err := h.refreshMetricCache(c.Request().Context()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to refresh metrics")
	}

	// Extract query context using OpenAI
	queryCtx, err := h.extractQueryContext(c.Request().Context(), req.Query)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process query")
	}

	// Build PromQL query
	promQL, warnings := h.buildPromQLQuery(queryCtx)
	if promQL == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Could not generate valid PromQL")
	}

	// Validate generated query
	validationResult, err := h.promClient.ValidateQuery(c.Request().Context(), promQL)
	if err != nil {
		fmt.Printf("Validation error: %v\n", err)
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Query validation error: %v", err))
	}
	if !validationResult.Valid {
		fmt.Printf("Invalid query: %s\nError: %s\n", promQL, validationResult.Error)
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid PromQL query: %s", validationResult.Error))
	}

	explanation := h.generateExplanation(queryCtx, promQL)

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

func (h *Handlers) extractQueryContext(ctx context.Context, query string) (*types.QueryContext, error) {
	// Build system context with available metrics
	metrics := make([]string, 0, len(h.metricCache))
	services := make(map[string][]string)

	for name, schema := range h.metricCache {
		metrics = append(metrics, name)
		if job, ok := schema.Labels["job"]; ok {
			services[job] = append(services[job], name)
		}
	}

	fmt.Printf("Discovered services:\n")
	for service, serviceMetrics := range services {
		fmt.Printf("Service %s metrics: %v\n", service, serviceMetrics)
	}

	systemContext := fmt.Sprintln(`You are a PromQL query builder. Return only a valid JSON object.
Available metrics and their labels:
prometheus_http_requests_total with labels: code=["200","302","400","404"], handler=[paths]

Return JSON with exact fields:
{
  "metric": "prometheus_http_requests_total",
  "labels": {"code": "400"},  // Use actual label values from metrics
  "timeRange": "5m",
  "aggregation": "rate"
}

IMPORTANT: ONLY return valid JSON, no explanation.`, strings.Join(metrics, "\n"))

	prompt := systemContext + "\n\nQuery: " + query
	fmt.Printf("Sending prompt to OpenAI:\n%s\n", prompt)

	result, err := h.openaiClient.Complete(ctx, prompt)
	if err != nil {
		fmt.Printf("OpenAI error: %v\n", err)
		return nil, err
	}

	var extracted struct {
		Metric      string            `json:"metric"`
		Labels      map[string]string `json:"labels"`
		TimeRange   string            `json:"timeRange"`
		Aggregation string            `json:"aggregation"`
	}

	fmt.Printf("OpenAI response:\n%s\n", result)

	// Try to find JSON object in response if there's surrounding text
	jsonStart := strings.Index(result, "{")
	jsonEnd := strings.LastIndex(result, "}")
	if jsonStart >= 0 && jsonEnd >= 0 && jsonEnd > jsonStart {
		result = result[jsonStart : jsonEnd+1]
	}

	if err := json.Unmarshal([]byte(result), &extracted); err != nil {
		fmt.Printf("JSON unmarshal error: %v\nAttempted to parse: %s\n", err, result)
		return nil, fmt.Errorf("invalid response format: %v", err)
	}

	// Parse TimeRange
	duration, err := time.ParseDuration(extracted.TimeRange)
	if err != nil {
		return nil, fmt.Errorf("invalid time range format: %v", err)
	}

	timeRange := types.TimeRange{
		Duration: duration,
	}

	return &types.QueryContext{
		Query:       query,
		MainMetric:  extracted.Metric,
		Labels:      extracted.Labels,
		TimeRange:   timeRange,
		Aggregation: extracted.Aggregation,
	}, nil
}

func (h *Handlers) buildPromQLQuery(ctx *types.QueryContext) (string, []string) {
	var warnings []string
	var parts []string

	if ctx.MainMetric == "" {
		warnings = append(warnings, "no metric specified")
		return "", warnings
	}

	// Add aggregation
	if ctx.Aggregation != "" {
		parts = append(parts, ctx.Aggregation)
		parts = append(parts, "(")
	}

	// Add metric name
	parts = append(parts, ctx.MainMetric)

	// Add labels if present
	if len(ctx.Labels) > 0 {
		labelParts := make([]string, 0, len(ctx.Labels))
		for k, v := range ctx.Labels {
			labelParts = append(labelParts, fmt.Sprintf("%s=%q", k, v))
		}
		parts = append(parts, "{"+strings.Join(labelParts, ",")+"}")
	}

	// Add time range if present
	if ctx.TimeRange.Duration > 0 {
		parts = append(parts, fmt.Sprintf("[%s]", ctx.TimeRange.Duration.String()))
	}

	// Close aggregation if present
	if ctx.Aggregation != "" {
		parts = append(parts, ")")
	}

	promQL := strings.Join(parts, "")
	fmt.Printf("Generated PromQL: %s\n", promQL)
	return promQL, warnings
}

func (h *Handlers) generateExplanation(ctx *types.QueryContext, promQL string) string {
	// Build explanation prompt
	prompt := fmt.Sprintf(`Explain this PromQL query in natural language:
Query: %s
Original question: %s
Metric: %s
Aggregation: %s
Labels: %v

Return a concise explanation in one sentence.`,
		promQL, ctx.Query, ctx.MainMetric, ctx.Aggregation, ctx.Labels)

	result, err := h.openaiClient.Complete(context.Background(), prompt)
	if err != nil {
		return fmt.Sprintf("Query: %s\nError generating explanation: %v", promQL, err)
	}

	// Clean up any markdown or quotes
	explanation := strings.TrimSpace(result)
	explanation = strings.Trim(explanation, "`\"")

	return explanation
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
