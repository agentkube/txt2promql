package parser

import (
	"regexp"
	"strings"
)

type Normalizer struct {
	replacements map[string]string
	patterns     map[string]*regexp.Regexp
}

func NewNormalizer() *Normalizer {
	return &Normalizer{
		replacements: map[string]string{
			"last hour":    "1h",
			"past hour":    "1h",
			"last day":     "24h",
			"past day":     "24h",
			"last week":    "7d",
			"past week":    "7d",
			"greater than": ">",
			"less than":    "<",
			"equal to":     "=",
		},
		patterns: map[string]*regexp.Regexp{
			"whitespace": regexp.MustCompile(`\s+`),
			"time":       regexp.MustCompile(`(\d+)\s*(hour|minute|second|day)s?`),
		},
	}
}

func (n *Normalizer) Normalize(query string) (string, error) {
	// Convert to lowercase
	normalized := strings.ToLower(query)

	// Replace common phrases
	for phrase, replacement := range n.replacements {
		normalized = strings.ReplaceAll(normalized, phrase, replacement)
	}

	// Normalize time expressions
	normalized = n.patterns["time"].ReplaceAllStringFunc(normalized, func(match string) string {
		var unit string
		switch {
		case strings.Contains(match, "hour"):
			unit = "h"
		case strings.Contains(match, "minute"):
			unit = "m"
		case strings.Contains(match, "second"):
			unit = "s"
		case strings.Contains(match, "day"):
			unit = "d"
		}
		num := n.patterns["time"].FindStringSubmatch(match)[1]
		return num + unit
	})

	// Normalize whitespace
	normalized = n.patterns["whitespace"].ReplaceAllString(normalized, " ")
	normalized = strings.TrimSpace(normalized)

	return normalized, nil
}
