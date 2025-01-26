package agent

import (
	"fmt"
	"strings"

	"github.com/agentkube/txt2promql/internal/types"
)

type QueryBuilder struct{}

func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{}
}

func (qb *QueryBuilder) Build(ctx *types.QueryContext) (string, []string) {
	var warnings []string
	var parts []string

	if ctx.MainMetric == "" {
		warnings = append(warnings, "no metric specified")
		return "", warnings
	}

	if ctx.Aggregation != "" {
		parts = append(parts, ctx.Aggregation, "(")
	}

	parts = append(parts, ctx.MainMetric)

	if len(ctx.Labels) > 0 {
		labelParts := make([]string, 0, len(ctx.Labels))
		for k, v := range ctx.Labels {
			labelParts = append(labelParts, fmt.Sprintf("%s=%q", k, v))
		}
		parts = append(parts, "{"+strings.Join(labelParts, ",")+"}")
	}

	if ctx.TimeRange.Duration > 0 {
		parts = append(parts, fmt.Sprintf("[%s]", ctx.TimeRange.Duration.String()))
	}

	if ctx.Aggregation != "" {
		parts = append(parts, ")")
	}

	promQL := strings.Join(parts, "")
	return promQL, warnings
}
