// internal/types/context.go
package types

import "time"

type QueryContext struct {
	Query         string
	Intent        string
	MainMetric    string
	Labels        map[string]string
	TimeRange     TimeRange
	Aggregation   string
	GroupBy       []string
	AdditionalOps []string
	Rules         []Rule
}

type TimeRange struct {
	Start    time.Time
	End      time.Time
	Duration time.Duration
}

type Rule struct {
	Pattern     string
	MetricType  string
	Labels      []string
	Aggregation string
}
