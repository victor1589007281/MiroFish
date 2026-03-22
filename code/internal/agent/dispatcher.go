package agent

import (
	"context"
	"swarm-predict/internal/model"
)

// Dispatcher assembles and runs the block pipeline for a given agent.
type Dispatcher struct {
	defaultBlocks []Block
}

func NewDispatcher(defaults ...Block) *Dispatcher {
	return &Dispatcher{defaultBlocks: defaults}
}

// Dispatch runs the full block pipeline and returns the final action proposal.
func (d *Dispatcher) Dispatch(ctx context.Context, ag *Agent, stepCtx *StepContext) (*model.ActionProposal, error) {
	allBlocks := append(d.defaultBlocks, ag.Blocks...)
	stepCtx.Profile = ag.Profile

	// Load social context if available
	if ag.Social != nil {
		rels, err := ag.Social.GetRelationships(ctx, ag.ID)
		if err == nil {
			stepCtx.SocialContext = rels
		}
	}

	var outputs []*BlockOutput
	for _, block := range allBlocks {
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

	return resolveAction(ag.ID, outputs), nil
}

// DispatchAll runs dispatch for multiple agents and returns all proposals.
func (d *Dispatcher) DispatchAll(ctx context.Context, agents []*Agent, stepCtx *StepContext) ([]model.ActionProposal, error) {
	var proposals []model.ActionProposal
	for _, ag := range agents {
		localCtx := *stepCtx
		proposal, err := d.Dispatch(ctx, ag, &localCtx)
		if err != nil {
			continue
		}
		if proposal != nil {
			proposals = append(proposals, *proposal)
		}
	}
	return proposals, nil
}

// resolveAction picks the best action from block outputs.
func resolveAction(agentID string, outputs []*BlockOutput) *model.ActionProposal {
	if len(outputs) == 0 {
		return &model.ActionProposal{
			AgentID:    agentID,
			Intent:     "observe",
			ActionType: "like",
		}
	}

	// Priority: posting/reaction output > planning > memory/reflection (context only)
	priorities := map[string]int{
		"posting":    100,
		"reaction":   90,
		"planning":   50,
		"social":     30,
		"reflection": 20,
		"memory":     10,
	}

	var bestOutput *BlockOutput
	bestPriority := -1
	for _, out := range outputs {
		p, ok := priorities[out.BlockName]
		if !ok {
			p = 0
		}
		if out.Details["action_type"] != "" && p > bestPriority {
			bestOutput = out
			bestPriority = p
		}
	}

	if bestOutput == nil {
		return &model.ActionProposal{
			AgentID:    agentID,
			Intent:     "observe",
			ActionType: "like",
		}
	}

	return &model.ActionProposal{
		AgentID:    agentID,
		Intent:     bestOutput.Intent,
		ActionType: bestOutput.Details["action_type"],
		Content:    bestOutput.Details["content"],
		TargetID:   bestOutput.Details["target_id"],
	}
}
