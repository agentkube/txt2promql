package knowledgegraph

import (
	"fmt"
	"strings"
	"sync"

	"github.com/agentkube/txt2promql/internal/prometheus"
)

type MetricPattern struct {
	Pattern     string
	TimeWindows []string
	Labels      []string
	Categories  []string
}

type KnowledgePatterns struct {
	patterns map[string][]MetricPattern
	mu       sync.RWMutex
}

func NewKnowledgePatterns() *KnowledgePatterns {
	kp := &KnowledgePatterns{
		patterns: make(map[string][]MetricPattern),
	}

	// Error rate patterns
	kp.AddPattern("error_rate", []MetricPattern{
		{
			Pattern:     "rate(errors[time]) / rate(total[time])",
			TimeWindows: []string{"5m", "1h", "24h"},
			Categories:  []string{"errors", "rates"},
		},
		{
			Pattern:     "sum(rate(errors{status=~'5..'}[time])) / sum(rate(total[time]))",
			TimeWindows: []string{"5m", "1h", "24h"},
			Categories:  []string{"errors", "http", "rates"},
		},
	})

	// Latency patterns
	kp.AddPattern("latency", []MetricPattern{
		{
			Pattern:     "histogram_quantile(0.95, rate(duration_bucket[time]))",
			TimeWindows: []string{"5m", "1h"},
			Categories:  []string{"latency", "performance"},
		},
		{
			Pattern:     "rate(duration_sum[time]) / rate(duration_count[time])",
			TimeWindows: []string{"5m", "1h"},
			Categories:  []string{"latency", "performance"},
		},
	})

	// Resource utilization patterns
	kp.AddPattern("utilization", []MetricPattern{
		{
			Pattern:     "sum by (instance) (rate(cpu_seconds_total[time]))",
			TimeWindows: []string{"5m", "1h"},
			Categories:  []string{"resources", "cpu"},
		},
		{
			Pattern:    "memory_used_bytes / memory_total_bytes * 100",
			Categories: []string{"resources", "memory"},
		},
	})

	return kp
}

func (kp *KnowledgePatterns) AddPattern(concept string, patterns []MetricPattern) {
	kp.mu.Lock()
	defer kp.mu.Unlock()
	kp.patterns[concept] = patterns
}

func (kp *KnowledgePatterns) FindPatterns(query string, availableMetrics map[string]struct{}) []MetricPattern {
	kp.mu.RLock()
	defer kp.mu.RUnlock()

	var matches []MetricPattern
	queryLower := strings.ToLower(query)

	// Check each concept's patterns
	for concept, patterns := range kp.patterns {
		if strings.Contains(queryLower, concept) {
			for _, pattern := range patterns {
				// Verify if required metrics exist
				hasRequiredMetrics := true
				for _, category := range pattern.Categories {
					found := false
					for metric := range availableMetrics {
						if strings.Contains(metric, category) {
							found = true
							break
						}
					}
					if !found {
						hasRequiredMetrics = false
						break
					}
				}
				if hasRequiredMetrics {
					matches = append(matches, pattern)
				}
			}
		}
	}

	return matches
}

type MetricInfo struct {
	Name        string            `json:"name"`
	Pattern     string            `json:"pattern"`
	Description string            `json:"description"`
	Labels      map[string]string `json:"labels,omitempty"`
}

func (kp *KnowledgePatterns) FindSimilarMetrics(metricName string, metricCache map[string]prometheus.MetricSchema) []MetricInfo {
	var similarMetrics []MetricInfo
	baseMetric := strings.TrimSuffix(metricName, "_count")
	suffixes := []string{"_count", "_sum", "_bucket"}

	originalMetric, exists := metricCache[metricName]
	if !exists {
		return nil
	}

	for name, schema := range metricCache {
		if name == metricName {
			continue
		}

		isRelated := false
		for _, suffix := range suffixes {
			if strings.HasPrefix(name, baseMetric) && strings.HasSuffix(name, suffix) {
				isRelated = true
				break
			}
		}

		commonLabels := 0
		for label := range originalMetric.Labels {
			if _, ok := schema.Labels[label]; ok {
				commonLabels++
			}
		}
		if commonLabels >= 2 {
			isRelated = true
		}

		if isRelated {
			pattern := buildPatternForMetric(name, schema.Labels)
			description := generateMetricDescription(name, baseMetric)

			similarMetrics = append(similarMetrics, MetricInfo{
				Name:        name,
				Pattern:     pattern,
				Description: description,
				Labels:      schema.Labels,
			})
		}

		if len(similarMetrics) >= 3 {
			break
		}
	}

	return similarMetrics
}

func buildPatternForMetric(metricName string, labels map[string]string) string {
	// Base pattern
	pattern := metricName

	// Common patterns based on suffix
	switch {
	case strings.HasSuffix(metricName, "_bucket"):
		pattern = fmt.Sprintf("histogram_quantile(0.95, rate(%s[5m]))", metricName)
	case strings.HasSuffix(metricName, "_count"):
		pattern = fmt.Sprintf("rate(%s[5m])", metricName)
	case strings.HasSuffix(metricName, "_sum"):
		pattern = fmt.Sprintf("%s / %s_count", metricName, strings.TrimSuffix(metricName, "_sum"))
	}

	// Add labels if present
	if len(labels) > 0 {
		labelPairs := make([]string, 0, len(labels))
		for k, v := range labels {
			if k != "__name__" {
				labelPairs = append(labelPairs, fmt.Sprintf("%s=\"%s\"", k, v))
			}
		}
		if len(labelPairs) > 0 {
			pattern += fmt.Sprintf("{%s}", strings.Join(labelPairs, ","))
		}
	}

	return pattern
}

func generateMetricDescription(metricName, baseMetric string) string {
	switch {
	case strings.HasSuffix(metricName, "_sum"):
		return fmt.Sprintf("Total sum for %s", baseMetric)
	case strings.HasSuffix(metricName, "_count"):
		return fmt.Sprintf("Request count for %s", baseMetric)
	case strings.HasSuffix(metricName, "_bucket"):
		return fmt.Sprintf("Duration buckets for calculating quantiles of %s", baseMetric)
	default:
		return fmt.Sprintf("Related metric: %s", metricName)
	}
}
