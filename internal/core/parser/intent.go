package parser

import (
	"regexp"
)

type Intent struct {
	Type      string
	Operation string
	TimeFrame string
	Modifiers map[string]string
}

type IntentParser struct {
	timePatterns map[string]*regexp.Regexp
	opPatterns   map[string]*regexp.Regexp
}

func NewIntentParser() *IntentParser {
	return &IntentParser{
		timePatterns: map[string]*regexp.Regexp{
			"last_hour": regexp.MustCompile(`(?i)last hour|past hour|1h`),
			"last_day":  regexp.MustCompile(`(?i)last day|past day|24h`),
			"last_week": regexp.MustCompile(`(?i)last week|past week|7d`),
		},
		opPatterns: map[string]*regexp.Regexp{
			"rate":  regexp.MustCompile(`(?i)rate|per second|velocity`),
			"avg":   regexp.MustCompile(`(?i)average|mean|avg`),
			"sum":   regexp.MustCompile(`(?i)sum|total`),
			"count": regexp.MustCompile(`(?i)count|number of`),
		},
	}
}

func (p *IntentParser) Parse(query string) (*Intent, error) {
	intent := &Intent{
		Type:      "instant",
		Modifiers: make(map[string]string),
	}

	// Detect time frame
	for frame, pattern := range p.timePatterns {
		if pattern.MatchString(query) {
			intent.TimeFrame = frame
			break
		}
	}

	// Detect operation
	for op, pattern := range p.opPatterns {
		if pattern.MatchString(query) {
			intent.Operation = op
			break
		}
	}

	// Determine type based on operation
	if intent.Operation == "rate" {
		intent.Type = "counter"
	}

	return intent, nil
}
