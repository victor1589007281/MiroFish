package blocks

import (
	"context"
	"fmt"
	"math/rand"

	"swarm-predict/internal/agent"
	"swarm-predict/internal/cognitive"
	"swarm-predict/internal/openclaw"
)

// PostingBlock decides whether and what to post on the social platform.
type PostingBlock struct {
	cc    openclaw.ChatCompleter
	model string
}

func NewPostingBlock(cc openclaw.ChatCompleter, llmModel string) *PostingBlock {
	return &PostingBlock{cc: cc, model: llmModel}
}

func (b *PostingBlock) Name() string { return "posting" }

func (b *PostingBlock) ShouldActivate(stepCtx *agent.StepContext) bool {
	return len(stepCtx.Feed) > 0 || rand.Float64() < 0.3
}

func (b *PostingBlock) Execute(ctx context.Context, stepCtx *agent.StepContext) (*agent.BlockOutput, error) {
	feedSummary := ""
	for i, post := range stepCtx.Feed {
		if i >= 5 {
			break
		}
		feedSummary += fmt.Sprintf("- @%s: %s\n", post.AuthorID, post.Content)
	}

	socialSummary := ""
	for _, rel := range stepCtx.SocialContext {
		socialSummary += fmt.Sprintf("- %s (like=%.1f)\n", rel.TargetID, rel.Likability)
	}

	prompt := fmt.Sprintf(`You are %s (%s). Personality: %s.
Your current plan: %s

Information feed:
%s

Your social relationships:
%s

Based on your personality, plan, and the current feed, decide your action.
Respond as JSON: {"action_type": "post|reply|like|repost|follow", "content": "...", "target_id": "post_id or user_id if applicable"}`,
		stepCtx.Profile.Name, stepCtx.Profile.Role, stepCtx.Profile.Personality,
		planText(stepCtx.Plan),
		feedSummary,
		socialSummary)

	resp, err := b.cc.Complete(ctx, openclaw.ChatRequest{
		Model: b.model,
		Messages: []openclaw.ChatMessage{
			{Role: "system", Content: "You are a social media user simulation agent. Always respond with valid JSON."},
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		return nil, err
	}

	var decision struct {
		ActionType string `json:"action_type"`
		Content    string `json:"content"`
		TargetID   string `json:"target_id"`
	}
	if err := cognitive.ExtractJSON(resp, &decision); err != nil {
		content := cognitive.ExtractContent(resp)
		return &agent.BlockOutput{
			BlockName: "posting",
			Intent:    content,
			Details: map[string]string{
				"action_type": "post",
				"content":     content,
			},
		}, nil
	}

	return &agent.BlockOutput{
		BlockName: "posting",
		Intent:    fmt.Sprintf("%s wants to %s", stepCtx.Profile.Name, decision.ActionType),
		Details: map[string]string{
			"action_type": decision.ActionType,
			"content":     decision.Content,
			"target_id":   decision.TargetID,
		},
	}, nil
}

func planText(p *cognitive.Plan) string {
	if p == nil {
		return "no specific plan"
	}
	return fmt.Sprintf("%s (current: %s)", p.DailyGoal, p.CurrentStep)
}
