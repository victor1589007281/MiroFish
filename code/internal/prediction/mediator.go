package prediction

import (
	"context"
	"fmt"
	"strings"

	"swarm-predict/internal/cognitive"
	"swarm-predict/internal/openclaw"
)

// Mediator synthesizes expert predictions and surfaces disagreements.
type Mediator struct {
	cc    openclaw.ChatCompleter
	model string
}

func NewMediator(cc openclaw.ChatCompleter, llmModel string) *Mediator {
	return &Mediator{cc: cc, model: llmModel}
}

// MediationResult contains the mediator's analysis.
type MediationResult struct {
	Feedback      string   `json:"feedback"`
	Disagreements []string `json:"disagreements"`
	KeyArguments  []string `json:"key_arguments"`
}

// Mediate analyzes expert predictions and generates structured feedback.
func (m *Mediator) Mediate(ctx context.Context, question string, expertPredictions map[string]ExpertOutput) (*MediationResult, error) {
	var sb strings.Builder
	for name, ep := range expertPredictions {
		fmt.Fprintf(&sb, "Expert '%s': probability=%.2f\nReasoning: %s\n\n", name, ep.Probability, ep.Reasoning)
	}

	prompt := fmt.Sprintf(`Question: %s

Expert predictions:
%s

Analyze the expert predictions:
1. Identify the key points of disagreement
2. Highlight the strongest arguments from each side
3. Provide structured feedback that would help experts refine their estimates

Output JSON: {"feedback": "...", "disagreements": ["..."], "key_arguments": ["..."]}`,
		question, sb.String())

	resp, err := m.cc.Complete(ctx, openclaw.ChatRequest{
		Model: m.model,
		Messages: []openclaw.ChatMessage{
			{Role: "system", Content: mediatorSystemPrompt},
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("mediate: %w", err)
	}

	var result MediationResult
	if err := cognitive.ExtractJSON(resp, &result); err != nil {
		result.Feedback = cognitive.ExtractContent(resp)
	}
	return &result, nil
}

// ExpertOutput holds a single expert's prediction.
type ExpertOutput struct {
	Probability float64 `json:"probability"`
	Reasoning   string  `json:"reasoning"`
}

const mediatorSystemPrompt = `You are a forecasting mediator. Your role is to:
1. Identify genuine disagreements vs semantic differences
2. Surface the strongest evidence and arguments from each expert
3. Provide feedback that helps experts refine their predictions toward accuracy
Always respond with valid JSON.`
