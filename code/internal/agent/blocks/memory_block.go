package blocks

import (
	"context"
	"fmt"
	"strings"

	"swarm-predict/internal/agent"
	"swarm-predict/internal/cognitive"
)

// MemoryBlock retrieves relevant memories and records observations.
type MemoryBlock struct {
	memory *cognitive.MemoryStream
	topK   int
}

func NewMemoryBlock(memory *cognitive.MemoryStream, topK int) *MemoryBlock {
	if topK <= 0 {
		topK = 5
	}
	return &MemoryBlock{memory: memory, topK: topK}
}

func (b *MemoryBlock) Name() string { return "memory" }

func (b *MemoryBlock) ShouldActivate(_ *agent.StepContext) bool { return true }

func (b *MemoryBlock) Execute(ctx context.Context, stepCtx *agent.StepContext) (*agent.BlockOutput, error) {
	query := buildQueryFromContext(stepCtx)

	memories, err := b.memory.Retrieve(ctx, query, b.topK)
	if err != nil {
		return nil, fmt.Errorf("retrieve memories: %w", err)
	}

	details := map[string]string{
		"retrieved_memories": cognitive.FormatMemories(memories),
	}

	for _, obs := range stepCtx.Observations {
		importance := 0.5
		if strings.Contains(strings.ToLower(obs), "breaking") || strings.Contains(strings.ToLower(obs), "important") {
			importance = 0.8
		}
		b.memory.Add(ctx, obs, "observation", importance)
	}

	return &agent.BlockOutput{
		BlockName: "memory",
		Intent:    "memory context loaded",
		Details:   details,
	}, nil
}

func buildQueryFromContext(stepCtx *agent.StepContext) string {
	parts := []string{stepCtx.Profile.Role}
	if stepCtx.Plan != nil {
		parts = append(parts, stepCtx.Plan.CurrentStep)
	}
	for _, obs := range stepCtx.Observations {
		if len(obs) > 100 {
			parts = append(parts, obs[:100])
		} else {
			parts = append(parts, obs)
		}
	}
	return strings.Join(parts, ". ")
}
