// internal/provider/openai/processor.go
package openai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/agentkube/txt2promql/internal/types"
	"github.com/agentkube/txt2promql/pkg/ai"
)

type Processor struct {
	openai_client *OpenAIClient
}

func NewProcessor(openai_client *OpenAIClient) *Processor {
	return &Processor{openai_client: openai_client}
}

func (p *Processor) ExtractContext(ctx context.Context, query string) (*types.QueryContext, error) {
	fmt.Println("ExtractContext query", query)

	prompt := fmt.Sprintf(ai.PromptMap["PromQLContextExtractor"], query)
	fmt.Println("prompt", prompt)

	result, err := p.openai_client.Complete(ctx, prompt)

	fmt.Println("result", result)

	if err != nil {
		return nil, err
	}

	var extracted struct {
		Metric      string            `json:"metric"`
		Labels      map[string]string `json:"labels"`
		TimeRange   string            `json:"timeRange"`
		Aggregation string            `json:"aggregation"`
	}

	if err := json.Unmarshal([]byte(result), &extracted); err != nil {
		return nil, err
	}

	return &types.QueryContext{
		MainMetric:  extracted.Metric,
		Labels:      extracted.Labels,
		Aggregation: extracted.Aggregation,
	}, nil
}
