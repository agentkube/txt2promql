package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/agentkube/txt2promql/internal/provider/openai"
	"github.com/agentkube/txt2promql/internal/types"
	"github.com/agentkube/txt2promql/pkg/ai"
)

type Explainer struct {
	openaiClient *openai.OpenAIClient
}

func NewExplainer(openaiClient *openai.OpenAIClient) *Explainer {
	return &Explainer{
		openaiClient: openaiClient,
	}
}

func (e *Explainer) GenerateExplanation(ctx context.Context, queryCtx *types.QueryContext, promQL string) string {
	prompt := fmt.Sprintf(ai.PromptMap["PromQLExplanation"],
		promQL, queryCtx.Query, queryCtx.MainMetric, queryCtx.Aggregation, queryCtx.Labels)

	result, err := e.openaiClient.Complete(ctx, prompt)
	if err != nil {
		return fmt.Sprintf("Query: %s\nError generating explanation: %v", promQL, err)
	}

	explanation := strings.TrimSpace(result)
	explanation = strings.Trim(explanation, "`\"")

	return explanation
}
