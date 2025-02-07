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
	"github.com/agentkube/txt2promql/pkg/ai"
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
	systemContext := fmt.Sprintf(ai.PromptMap["PromQLBuilder"], strings.Join(metricsDescription, "\n"))

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
