package blocks

import (
	"context"
	"time"

	"swarm-predict/internal/agent"
	"swarm-predict/internal/cognitive"
	"swarm-predict/internal/openclaw"
)

// ReflectionBlock triggers the reflection engine when importance accumulates.
type ReflectionBlock struct {
	engine         *cognitive.ReflectionEngine
	memory         *cognitive.MemoryStream
	lastReflection time.Time
}

func NewReflectionBlock(cc openclaw.ChatCompleter, memory *cognitive.MemoryStream, llmModel string) *ReflectionBlock {
	return &ReflectionBlock{
		engine:         cognitive.NewReflectionEngine(cc, llmModel),
		memory:         memory,
		lastReflection: time.Now().Add(-1 * time.Hour),
	}
}

func (b *ReflectionBlock) Name() string { return "reflection" }

func (b *ReflectionBlock) ShouldActivate(stepCtx *agent.StepContext) bool {
	return stepCtx.Round%3 == 0 // reflect every 3 rounds
}

func (b *ReflectionBlock) Execute(ctx context.Context, stepCtx *agent.StepContext) (*agent.BlockOutput, error) {
	insights, err := b.engine.MaybeReflect(ctx, b.memory, b.lastReflection)
	if err != nil {
		return nil, err
	}
	if insights == nil {
		return nil, nil
	}

	b.lastReflection = time.Now()

	insightSummary := ""
	for _, ins := range insights {
		insightSummary += ins.Content + "; "
	}

	return &agent.BlockOutput{
		BlockName: "reflection",
		Intent:    "reflected on recent experiences",
		Details: map[string]string{
			"insights": insightSummary,
		},
	}, nil
}
