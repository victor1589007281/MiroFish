package cognitive

import (
	"context"
	"fmt"

	"swarm-predict/internal/model"
	"swarm-predict/internal/openclaw"
)

// PlanningEngine generates hierarchical plans for agents.
type PlanningEngine struct {
	cc    openclaw.ChatCompleter
	model string
}

func NewPlanningEngine(cc openclaw.ChatCompleter, llmModel string) *PlanningEngine {
	return &PlanningEngine{cc: cc, model: llmModel}
}

// Plan represents a hierarchical plan decomposition.
type Plan struct {
	DailyGoal   string   `json:"daily_goal"`
	HourlySteps []string `json:"hourly_steps"`
	CurrentStep string   `json:"current_step"`
}

// GeneratePlan creates a plan based on agent profile, memories, and current context.
func (p *PlanningEngine) GeneratePlan(ctx context.Context, profile model.AgentProfile, memories []model.MemoryEntry, currentContext string) (*Plan, error) {
	prompt := fmt.Sprintf(`Agent profile:
Name: %s
Role: %s
Personality: %s
Goals: %v

Recent memories:
%s

Current situation: %s`,
		profile.Name, profile.Role, profile.Personality, profile.Goals,
		FormatMemories(memories), currentContext)

	resp, err := p.cc.Complete(ctx, openclaw.ChatRequest{
		Model: p.model,
		Messages: []openclaw.ChatMessage{
			{Role: "system", Content: planningSystemPrompt},
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("generate plan: %w", err)
	}

	var plan Plan
	if err := ExtractJSON(resp, &plan); err != nil {
		return &Plan{
			DailyGoal:   ExtractContent(resp),
			CurrentStep: "observe and react",
		}, nil
	}
	return &plan, nil
}

// AdjustPlan updates the current plan when an unexpected event occurs.
func (p *PlanningEngine) AdjustPlan(ctx context.Context, currentPlan *Plan, event string, profile model.AgentProfile) (*Plan, error) {
	prompt := fmt.Sprintf(`Current plan:
Daily goal: %s
Current step: %s

Unexpected event: %s

Agent: %s (%s)

Adjust the plan to account for this event. Output JSON with daily_goal, hourly_steps, current_step.`,
		currentPlan.DailyGoal, currentPlan.CurrentStep, event, profile.Name, profile.Role)

	resp, err := p.cc.Complete(ctx, openclaw.ChatRequest{
		Model: p.model,
		Messages: []openclaw.ChatMessage{
			{Role: "system", Content: planAdjustSystemPrompt},
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		return currentPlan, nil
	}

	var adjusted Plan
	if err := ExtractJSON(resp, &adjusted); err != nil {
		return currentPlan, nil
	}
	return &adjusted, nil
}

const planningSystemPrompt = `You are a planning module for a simulated social agent. Based on the agent's profile, memories, and current situation, generate a plan as JSON:
{"daily_goal": "...", "hourly_steps": ["step1", "step2", ...], "current_step": "immediate next action"}
The plan should be consistent with the agent's personality and goals.`

const planAdjustSystemPrompt = `You are adjusting a simulated agent's plan due to an unexpected event. Maintain the agent's personality and goals while adapting to the new situation. Output JSON: {"daily_goal": "...", "hourly_steps": [...], "current_step": "..."}`
