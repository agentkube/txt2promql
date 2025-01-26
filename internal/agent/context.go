package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	parser "github.com/agentkube/txt2promql/internal/core/parser"
	prometheus "github.com/agentkube/txt2promql/internal/prometheus"
	openai "github.com/agentkube/txt2promql/internal/provider/openai"
	types "github.com/agentkube/txt2promql/internal/types"
)

type ContextExtractor struct {
	openaiClient *openai.OpenAIClient
	intentParser *parser.IntentParser
	nerParser    *parser.NERParser
	normalizer   *parser.Normalizer
}

func NewContextExtractor(openaiClient *openai.OpenAIClient) *ContextExtractor {
	return &ContextExtractor{
		openaiClient: openaiClient,
		intentParser: parser.NewIntentParser(),
		nerParser:    parser.NewNERParser(),
		normalizer:   parser.NewNormalizer(),
	}
}

func (ce *ContextExtractor) ExtractQueryContext(ctx context.Context, query string, metrics map[string]prometheus.MetricSchema) (*types.QueryContext, error) {
	// Build metric info maps
	metricInfo := make(map[string]map[string]map[string]int)
	services := make(map[string][]string)

	// Process metrics and their labels
	for name, schema := range metrics {
		if job, ok := schema.Labels["job"]; ok {
			services[job] = append(services[job], name)
		}

		metricInfo[name] = make(map[string]map[string]int)
		for label, value := range schema.Labels {
			if label == "__name__" {
				continue
			}
			if _, ok := metricInfo[name][label]; !ok {
				metricInfo[name][label] = make(map[string]int)
			}
			metricInfo[name][label][value]++
		}
	}

	// Print services and metrics
	fmt.Printf("\nDiscovered Services and Metrics:\n")
	for service, serviceMetrics := range services {
		fmt.Printf("Service %s metrics:\n", service)
		for _, metric := range serviceMetrics {
			fmt.Printf("  - %s\n", metric)
		}
	}

	// Format metric descriptions
	var metricsDescription []string
	for metricName, labels := range metricInfo {
		labelInfo := make([]string, 0)
		for label, values := range labels {
			valueList := make([]string, 0)
			for value := range values {
				valueList = append(valueList, value)
			}
			labelInfo = append(labelInfo, fmt.Sprintf("%s=[%s]", label, strings.Join(valueList, ", ")))
		}
		metricStr := metricName
		if len(labelInfo) > 0 {
			metricStr += fmt.Sprintf(" with labels: %s", strings.Join(labelInfo, ", "))
		}
		metricsDescription = append(metricsDescription, metricStr)
	}

	fmt.Printf("\nMetric Details:\n%s\n", strings.Join(metricsDescription, "\n"))

	// Build system context with examples
	systemContext := fmt.Sprintf(`You are a PromQL query builder.

Available metrics and their labels:
%s

Valid PromQL examples:
- Simple sum: sum(prometheus_http_response_size_bytes_sum)
- Rate with time: rate(prometheus_http_requests_total[5m])
- Filtered sum: sum(prometheus_http_requests_total{code="200"})

Return ONLY a JSON object with these fields:
{
  "metric": "exact_metric_name_from_list",
  "labels": {"label": "value"},
  "timeRange": "5m",      // Omit for sum aggregation
  "aggregation": ""    // sum/rate/avg/count/increase - Leave empty if no aggregation needed
}

Rules:
1. Exact metric names only
2. Only use existing label values
3. Omit timeRange for sum operations
4. Use appropriate aggregation:
   - sum: for totals and sizes
   - rate: for per-second metrics
   - avg: for averages
   - count: for occurrences
   - increase: for total increases`, strings.Join(metricsDescription, "\n"))

	// Process query
	prompt := systemContext + "\n\nQuery: " + query
	fmt.Printf("\nProcessing query: %s\n", query)

	result, err := ce.openaiClient.Complete(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("OpenAI error: %w", err)
	}

	fmt.Printf("\nAgent Response:\n%s\n", result)

	// Extract JSON from response
	jsonStart := strings.Index(result, "{")
	jsonEnd := strings.LastIndex(result, "}")
	if jsonStart >= 0 && jsonEnd >= 0 && jsonEnd > jsonStart {
		result = result[jsonStart : jsonEnd+1]
	}

	var extracted struct {
		Metric      string            `json:"metric"`
		Labels      map[string]string `json:"labels"`
		TimeRange   string            `json:"timeRange"`
		Aggregation string            `json:"aggregation"`
	}

	if err := json.Unmarshal([]byte(result), &extracted); err != nil {
		return nil, fmt.Errorf("invalid response format: %w", err)
	}

	// Parse timeRange if present
	var timeRange types.TimeRange
	if extracted.TimeRange != "" {
		duration, err := time.ParseDuration(extracted.TimeRange)
		if err != nil {
			return nil, fmt.Errorf("invalid time range format: %w", err)
		}
		timeRange.Duration = duration
	}

	return &types.QueryContext{
		Query:       query,
		MainMetric:  extracted.Metric,
		Labels:      extracted.Labels,
		TimeRange:   timeRange,
		Aggregation: extracted.Aggregation,
	}, nil
}
