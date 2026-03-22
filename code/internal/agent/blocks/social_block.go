package blocks

import (
	"context"
	"fmt"
	"strings"

	"swarm-predict/internal/agent"
	"swarm-predict/internal/cognitive"
	"swarm-predict/internal/openclaw"
)

// SocialBlock evaluates interactions and updates social perception.
type SocialBlock struct {
	cc     openclaw.ChatCompleter
	social *cognitive.SocialGraph
	model  string
}

func NewSocialBlock(cc openclaw.ChatCompleter, social *cognitive.SocialGraph, llmModel string) *SocialBlock {
	return &SocialBlock{cc: cc, social: social, model: llmModel}
}

func (b *SocialBlock) Name() string { return "social" }

func (b *SocialBlock) ShouldActivate(stepCtx *agent.StepContext) bool {
	return len(stepCtx.Feed) > 0
}

func (b *SocialBlock) Execute(ctx context.Context, stepCtx *agent.StepContext) (*agent.BlockOutput, error) {
	interactions := extractInteractions(stepCtx)
	if len(interactions) == 0 {
		return nil, nil
	}

	prompt := fmt.Sprintf(`You are %s (%s). Evaluate how the following interactions affect your perception of each person.
For each person, provide a likability_delta and reputation_delta (each between -0.3 and 0.3).
Output JSON array: [{"target_id": "...", "likability_delta": 0.1, "reputation_delta": -0.1}]

Interactions:
%s`, stepCtx.Profile.Name, stepCtx.Profile.Role, strings.Join(interactions, "\n"))

	resp, err := b.cc.Complete(ctx, openclaw.ChatRequest{
		Model: b.model,
		Messages: []openclaw.ChatMessage{
			{Role: "system", Content: "You evaluate social interactions for a simulated agent. Always respond with valid JSON."},
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		return nil, err
	}

	var deltas []struct {
		TargetID   string  `json:"target_id"`
		LikeDelta  float64 `json:"likability_delta"`
		RepDelta   float64 `json:"reputation_delta"`
	}
	if err := cognitive.ExtractJSON(resp, &deltas); err == nil {
		for _, d := range deltas {
			b.social.UpdateAfterInteraction(ctx, stepCtx.Profile.ID, d.TargetID, cognitive.InteractionDelta{
				LikabilityDelta: d.LikeDelta,
				ReputationDelta: d.RepDelta,
			})
		}
	}

	return &agent.BlockOutput{
		BlockName: "social",
		Intent:    "social perception updated",
		Details:   map[string]string{},
	}, nil
}

func extractInteractions(stepCtx *agent.StepContext) []string {
	var result []string
	for _, post := range stepCtx.Feed {
		result = append(result, fmt.Sprintf("@%s posted: %s", post.AuthorID, post.Content))
		for _, reply := range post.Replies {
			result = append(result, fmt.Sprintf("  @%s replied: %s", reply.AuthorID, reply.Content))
		}
	}
	return result
}
