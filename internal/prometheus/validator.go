// internal/prometheus/validator.go
package prometheus

import (
	"context"
	"fmt"
	"strings"
)

type ValidationResult struct {
	Valid    bool     `json:"valid"`
	Warnings []string `json:"warnings"`
	Error    string   `json:"error,omitempty"`
}

func (c *Client) ValidateQuery(ctx context.Context, query string) (*ValidationResult, error) {
	result := &ValidationResult{Valid: true}

	if strings.TrimSpace(query) == "" {
		result.Valid = false
		result.Error = "empty query"
		return result, nil
	}

	_, err := c.Query(ctx, query)
	if err != nil {
		result.Valid = false
		result.Error = fmt.Sprintf("invalid query: %v", err)
		return result, nil
	}

	if len(query) > 1000 {
		result.Warnings = append(result.Warnings, "query exceeds recommended length")
	}

	if strings.Count(query, "{") != strings.Count(query, "}") {
		result.Warnings = append(result.Warnings, "unbalanced curly braces")
	}

	return result, nil
}
