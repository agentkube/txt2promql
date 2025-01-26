package parser

// NameEntityExtractor
import (
	"regexp"
)

type Entity struct {
	Type  string
	Value string
	Start int
	End   int
}

type NERParser struct {
	metricPattern *regexp.Regexp
	labelPattern  *regexp.Regexp
	timePattern   *regexp.Regexp
	numberPattern *regexp.Regexp
}

func NewNERParser() *NERParser {
	return &NERParser{
		metricPattern: regexp.MustCompile(`\b[a-zA-Z_:][a-zA-Z0-9_:]*\b`),
		labelPattern:  regexp.MustCompile(`\b([a-zA-Z_][a-zA-Z0-9_]*)\s*=\s*["']?([^"'}\s]+)["']?\b`),
		timePattern:   regexp.MustCompile(`\b(\d+[smhd])\b`),
		numberPattern: regexp.MustCompile(`\b\d+(?:\.\d+)?\b`),
	}
}

func (p *NERParser) ExtractEntities(query string) ([]Entity, error) {
	var entities []Entity

	// Find metrics
	for _, match := range p.metricPattern.FindAllStringIndex(query, -1) {
		entities = append(entities, Entity{
			Type:  "metric",
			Value: query[match[0]:match[1]],
			Start: match[0],
			End:   match[1],
		})
	}

	// Find labels
	for _, match := range p.labelPattern.FindAllStringSubmatchIndex(query, -1) {
		entities = append(entities, Entity{
			Type:  "label",
			Value: query[match[2]:match[3]] + "=" + query[match[4]:match[5]],
			Start: match[0],
			End:   match[1],
		})
	}

	// Find time ranges
	for _, match := range p.timePattern.FindAllStringIndex(query, -1) {
		entities = append(entities, Entity{
			Type:  "time",
			Value: query[match[0]:match[1]],
			Start: match[0],
			End:   match[1],
		})
	}

	return entities, nil
}
