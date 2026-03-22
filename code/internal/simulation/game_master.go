package simulation

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"swarm-predict/internal/cognitive"
	"swarm-predict/internal/model"
	"swarm-predict/internal/openclaw"

	"golang.org/x/sync/errgroup"
)

// GameMaster arbitrates agent actions for plausibility and consistency.
type GameMaster struct {
	cc    openclaw.ChatCompleter
	model string
}

func NewGameMaster(cc openclaw.ChatCompleter, llmModel string) *GameMaster {
	return &GameMaster{cc: cc, model: llmModel}
}

// Arbitrate checks whether an action proposal is consistent with the agent's persona.
// Simple actions (like, follow, repost) are auto-approved via rules engine.
// Complex actions (post, reply) go through LLM validation.
func (gm *GameMaster) Arbitrate(ctx context.Context, proposal model.ActionProposal, profile model.AgentProfile, memories []model.MemoryEntry) (*model.GMVerdict, error) {
	if !gm.requiresLLM(proposal.ActionType) {
		return &model.GMVerdict{Approved: true, Reason: "auto-approved simple action"}, nil
	}

	prompt := fmt.Sprintf(`You are the Game Master of a social simulation. An agent proposes an action. Evaluate if it is consistent with the agent's persona and current context.

Agent: %s
Role: %s
Personality: %s
Background: %s

Proposed action: %s
Action type: %s
Content: %s

Recent memories:
%s

Respond with JSON: {"approved": true/false, "reason": "...", "suggestion": "alternative if rejected"}`,
		profile.Name, profile.Role, profile.Personality, profile.Background,
		proposal.Intent, proposal.ActionType, proposal.Content,
		cognitive.FormatMemories(memories))

	resp, err := gm.cc.Complete(ctx, openclaw.ChatRequest{
		Model: gm.model,
		Messages: []openclaw.ChatMessage{
			{Role: "system", Content: gmSystemPrompt},
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		return &model.GMVerdict{Approved: true, Reason: "LLM error, auto-approved"}, nil
	}

	var verdict struct {
		Approved   bool   `json:"approved"`
		Reason     string `json:"reason"`
		Suggestion string `json:"suggestion"`
	}
	if err := cognitive.ExtractJSON(resp, &verdict); err != nil {
		content := strings.ToLower(cognitive.ExtractContent(resp))
		approved := !strings.Contains(content, "reject") && !strings.Contains(content, "not consistent")
		return &model.GMVerdict{Approved: approved, Reason: cognitive.ExtractContent(resp)}, nil
	}

	result := &model.GMVerdict{
		Approved: verdict.Approved,
		Reason:   verdict.Reason,
	}
	if !verdict.Approved && verdict.Suggestion != "" {
		result.ModifiedAction = &model.ActionProposal{
			AgentID:    proposal.AgentID,
			Intent:     verdict.Suggestion,
			ActionType: proposal.ActionType,
			Content:    verdict.Suggestion,
		}
	}
	return result, nil
}

// ArbitrateBatch processes multiple proposals concurrently.
// Simple actions are auto-approved immediately; complex actions are sent to
// LLM arbitration in parallel.
func (gm *GameMaster) ArbitrateBatch(ctx context.Context, proposals []model.ActionProposal, profiles map[string]model.AgentProfile, memories map[string][]model.MemoryEntry) ([]model.Decision, error) {
	var (
		mu        sync.Mutex
		decisions []model.Decision
	)

	// Split simple (auto-approve) from complex (need LLM)
	var complexProposals []model.ActionProposal
	for _, p := range proposals {
		if !gm.requiresLLM(p.ActionType) {
			decisions = append(decisions, model.Decision{
				AgentID:    p.AgentID,
				ActionType: p.ActionType,
				Content:    p.Content,
				TargetID:   p.TargetID,
			})
		} else {
			complexProposals = append(complexProposals, p)
		}
	}

	if len(complexProposals) == 0 {
		return decisions, nil
	}

	// Concurrently arbitrate complex proposals via LLM
	g, gctx := errgroup.WithContext(ctx)
	for _, p := range complexProposals {
		p := p
		g.Go(func() error {
			profile := profiles[p.AgentID]
			mem := memories[p.AgentID]

			verdict, err := gm.Arbitrate(gctx, p, profile, mem)
			if err != nil {
				return nil
			}

			mu.Lock()
			defer mu.Unlock()

			if verdict.Approved {
				decisions = append(decisions, model.Decision{
					AgentID:    p.AgentID,
					ActionType: p.ActionType,
					Content:    p.Content,
					TargetID:   p.TargetID,
				})
			} else if verdict.ModifiedAction != nil {
				decisions = append(decisions, model.Decision{
					AgentID:    verdict.ModifiedAction.AgentID,
					ActionType: verdict.ModifiedAction.ActionType,
					Content:    verdict.ModifiedAction.Content,
				})
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return decisions, fmt.Errorf("concurrent GM arbitration: %w", err)
	}

	return decisions, nil
}

func (gm *GameMaster) requiresLLM(actionType string) bool {
	switch actionType {
	case "like", "follow", "repost":
		return false
	default:
		return true
	}
}

const gmSystemPrompt = `You are the Game Master of a social media simulation. Your role is to ensure agent behaviors are consistent with their personas. Evaluate proposed actions for plausibility. Simple actions like liking or following are always fine. For content creation (posts, replies), check if the tone, topic, and stance match the agent's personality and background. Always respond with JSON.`
