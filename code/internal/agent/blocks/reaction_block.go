package blocks

import (
	"context"
	"fmt"
	"math/rand"

	"swarm-predict/internal/agent"
	"swarm-predict/internal/cognitive"
	"swarm-predict/internal/openclaw"
)

// ReactionBlock handles reactive interactions: like, repost, follow, reply to existing content.
type ReactionBlock struct {
	cc    openclaw.ChatCompleter
	model string
}

func NewReactionBlock(cc openclaw.ChatCompleter, llmModel string) *ReactionBlock {
	return &ReactionBlock{cc: cc, model: llmModel}
}

func (b *ReactionBlock) Name() string { return "reaction" }

func (b *ReactionBlock) ShouldActivate(stepCtx *agent.StepContext) bool {
	return len(stepCtx.Feed) > 0
}

func (b *ReactionBlock) Execute(ctx context.Context, stepCtx *agent.StepContext) (*agent.BlockOutput, error) {
	if len(stepCtx.Feed) == 0 {
		return nil, nil
	}

	// Pick a post to react to (simple relevance: random for now, could be improved)
	targetPost := stepCtx.Feed[rand.Intn(len(stepCtx.Feed))]

	prompt := fmt.Sprintf(`You are %s (%s). Personality: %s.
You saw this post by @%s: "%s" (likes: %d)

How do you react? Choose ONE action:
- "like" if you agree or find it interesting
- "reply" with a brief comment if you have something to say
- "repost" if you want to share it with your followers
- "ignore" if it doesn't interest you

Respond as JSON: {"action": "like|reply|repost|ignore", "content": "reply text if applicable"}`,
		stepCtx.Profile.Name, stepCtx.Profile.Role, stepCtx.Profile.Personality,
		targetPost.AuthorID, targetPost.Content, targetPost.Likes)

	resp, err := b.cc.Complete(ctx, openclaw.ChatRequest{
		Model: b.model,
		Messages: []openclaw.ChatMessage{
			{Role: "system", Content: "You are a social media user. React naturally based on your personality. Always respond with valid JSON."},
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		return nil, err
	}

	var decision struct {
		Action  string `json:"action"`
		Content string `json:"content"`
	}
	if err := cognitive.ExtractJSON(resp, &decision); err != nil {
		return nil, nil
	}

	if decision.Action == "ignore" {
		return nil, nil
	}

	return &agent.BlockOutput{
		BlockName: "reaction",
		Intent:    fmt.Sprintf("react to @%s's post: %s", targetPost.AuthorID, decision.Action),
		Details: map[string]string{
			"action_type": decision.Action,
			"content":     decision.Content,
			"target_id":   targetPost.ID,
		},
	}, nil
}
