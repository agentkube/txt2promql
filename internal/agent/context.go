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
	// Normalize query
	normalizedQuery, err := ce.normalizer.Normalize(query)
	if err != nil {
		return nil, fmt.Errorf("query normalization failed: %w", err)
	}

	// Extract intent
	intent, err := ce.intentParser.Parse(normalizedQuery)
	if err != nil {
		return nil, fmt.Errorf("intent parsing failed: %w", err)
	}

	// Extract entities
	entities, err := ce.nerParser.ExtractEntities(normalizedQuery)
	if err != nil {
		return nil, fmt.Errorf("entity extraction failed: %w", err)
	}

	// Build system context with available metrics
	metricNames := make([]string, 0, len(metrics))
	services := make(map[string][]string)

	for name, schema := range metrics {
		metricNames = append(metricNames, name)
		if job, ok := schema.Labels["job"]; ok {
			services[job] = append(services[job], name)
		}
	}

	// Build metric context from entities
	var selectedMetric string
	labels := make(map[string]string)
	timeRange := "5m" // default

	for _, entity := range entities {
		switch entity.Type {
		case "metric":
			if _, exists := metrics[entity.Value]; exists {
				selectedMetric = entity.Value
			}
		case "label":
			parts := strings.Split(entity.Value, "=")
			if len(parts) == 2 {
				labels[parts[0]] = parts[1]
			}
		case "time":
			timeRange = entity.Value
		}
	}

	// If no metric found from NER, use OpenAI
	if selectedMetric == "" {
		systemContext := fmt.Sprintln(`You are a PromQL query builder. Return only a valid JSON object.
Available metrics and their labels:
prometheus_http_requests_total with labels: code=["200","302","400","404"], handler=[paths]

Return JSON with exact fields:
{
"metric": "prometheus_http_requests_total",
"labels": {"code": "400"},
"timeRange": "5m",
"aggregation": "rate"
}

IMPORTANT: ONLY return valid JSON, no explanation.`, strings.Join(metricNames, "\n"))

		prompt := systemContext + "\n\nQuery: " + normalizedQuery
		result, err := ce.openaiClient.Complete(ctx, prompt)
		if err != nil {
			return nil, fmt.Errorf("OpenAI error: %w", err)
		}

		var extracted struct {
			Metric      string            `json:"metric"`
			Labels      map[string]string `json:"labels"`
			TimeRange   string            `json:"timeRange"`
			Aggregation string            `json:"aggregation"`
		}

		jsonStart := strings.Index(result, "{")
		jsonEnd := strings.LastIndex(result, "}")
		if jsonStart >= 0 && jsonEnd >= 0 && jsonEnd > jsonStart {
			result = result[jsonStart : jsonEnd+1]
		}

		if err := json.Unmarshal([]byte(result), &extracted); err != nil {
			return nil, fmt.Errorf("invalid response format: %w", err)
		}

		selectedMetric = extracted.Metric
		// Merge OpenAI labels with NER labels
		for k, v := range extracted.Labels {
			if _, exists := labels[k]; !exists {
				labels[k] = v
			}
		}
		timeRange = extracted.TimeRange
	}

	duration, err := time.ParseDuration(timeRange)
	if err != nil {
		return nil, fmt.Errorf("invalid time range format: %w", err)
	}

	return &types.QueryContext{
		Query:       query,
		MainMetric:  selectedMetric,
		Labels:      labels,
		TimeRange:   types.TimeRange{Duration: duration},
		Aggregation: intent.Operation,
	}, nil
}
