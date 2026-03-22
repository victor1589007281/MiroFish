package blocks

import (
	"context"
	"fmt"

	"swarm-predict/internal/agent"
	"swarm-predict/internal/cognitive"
	"swarm-predict/internal/openclaw"
)

// PlanningBlock generates and adjusts hierarchical plans.
type PlanningBlock struct {
	engine *cognitive.PlanningEngine
	memory *cognitive.MemoryStream
}

func NewPlanningBlock(cc openclaw.ChatCompleter, memory *cognitive.MemoryStream, llmModel string) *PlanningBlock {
	return &PlanningBlock{
		engine: cognitive.NewPlanningEngine(cc, llmModel),
		memory: memory,
	}
}

func (b *PlanningBlock) Name() string { return "planning" }

func (b *PlanningBlock) ShouldActivate(stepCtx *agent.StepContext) bool {
	return stepCtx.Plan == nil || stepCtx.Round%5 == 1
}

func (b *PlanningBlock) Execute(ctx context.Context, stepCtx *agent.StepContext) (*agent.BlockOutput, error) {
	memories, _ := b.memory.Retrieve(ctx, stepCtx.Profile.Role+" goals", 5)

	var currentContext string
	if len(stepCtx.Observations) > 0 {
		currentContext = stepCtx.Observations[0]
	}

	if stepCtx.Plan != nil && len(stepCtx.Observations) > 0 {
		adjusted, err := b.engine.AdjustPlan(ctx, stepCtx.Plan, stepCtx.Observations[0], stepCtx.Profile)
		if err == nil {
			stepCtx.Plan = adjusted
			return &agent.BlockOutput{
				BlockName: "planning",
				Intent:    fmt.Sprintf("adjusted plan: %s", adjusted.CurrentStep),
				Details: map[string]string{
					"daily_goal":   adjusted.DailyGoal,
					"current_step": adjusted.CurrentStep,
				},
			}, nil
		}
	}

	plan, err := b.engine.GeneratePlan(ctx, stepCtx.Profile, memories, currentContext)
	if err != nil {
		return nil, err
	}
	stepCtx.Plan = plan

	return &agent.BlockOutput{
		BlockName: "planning",
		Intent:    fmt.Sprintf("planned: %s", plan.CurrentStep),
		Details: map[string]string{
			"daily_goal":   plan.DailyGoal,
			"current_step": plan.CurrentStep,
		},
	}, nil
}
