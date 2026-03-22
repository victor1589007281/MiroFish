package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"swarm-predict/internal/cognitive"
	"swarm-predict/internal/model"
)

// StepContext carries the per-step information available to blocks.
type StepContext struct {
	Round         int
	Feed          []model.Post
	Profile       model.AgentProfile
	Observations  []string
	Plan          *cognitive.Plan
	SocialContext []model.SocialRelation
}

// BlockOutput is the result of a block's execution.
type BlockOutput struct {
	BlockName string
	Intent    string
	Details   map[string]string
}

// Block is the interface for pluggable agent capability modules.
type Block interface {
	Name() string
	ShouldActivate(ctx *StepContext) bool
	Execute(ctx context.Context, stepCtx *StepContext) (*BlockOutput, error)
}

// Agent is the central entity with a pluggable block architecture.
type Agent struct {
	ID                 string
	Profile            model.AgentProfile
	Blocks             []Block
	Memory             *cognitive.MemoryStream
	Social             *cognitive.SocialGraph
	LastReflectionTime time.Time
}

func NewAgent(id string, profile model.AgentProfile, memory *cognitive.MemoryStream, social *cognitive.SocialGraph, blocks ...Block) *Agent {
	return &Agent{
		ID:                 id,
		Profile:            profile,
		Blocks:             blocks,
		Memory:             memory,
		Social:             social,
		LastReflectionTime: time.Now(),
	}
}

// Step runs all activated blocks and uses priority-based resolution to pick the best action.
func (a *Agent) Step(ctx context.Context, stepCtx *StepContext) (*model.ActionProposal, error) {
	stepCtx.Profile = a.Profile

	// Load social context if available
	if a.Social != nil {
		rels, err := a.Social.GetRelationships(ctx, a.ID)
		if err == nil {
			stepCtx.SocialContext = rels
		}
	}

	var outputs []*BlockOutput
	for _, block := range a.Blocks {
		if block.ShouldActivate(stepCtx) {
			out, err := block.Execute(ctx, stepCtx)
			if err != nil {
				continue
			}
			if out != nil {
				outputs = append(outputs, out)
			}
		}
	}

	return resolveAction(a.ID, outputs), nil
}

// FormatForPrompt produces a text summary of the agent state for LLM context.
func (a *Agent) FormatForPrompt(stepCtx *StepContext) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Agent: %s\nRole: %s\nPersonality: %s\n", a.Profile.Name, a.Profile.Role, a.Profile.Personality)
	if stepCtx.Plan != nil {
		fmt.Fprintf(&sb, "Current plan: %s\nCurrent step: %s\n", stepCtx.Plan.DailyGoal, stepCtx.Plan.CurrentStep)
	}
	if len(stepCtx.SocialContext) > 0 {
		sb.WriteString("Social relationships:\n")
		for _, rel := range stepCtx.SocialContext {
			fmt.Fprintf(&sb, "  - %s: likability=%.1f, reputation=%.1f\n", rel.TargetID, rel.Likability, rel.Reputation)
		}
	}
	return sb.String()
}
