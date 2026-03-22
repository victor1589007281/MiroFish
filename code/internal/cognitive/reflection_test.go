package cognitive

import (
	"context"
	"testing"
	"time"

	"swarm-predict/internal/openclaw"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReflectionEngine_BelowThreshold(t *testing.T) {
	mockCC := openclaw.NewMockChatClient()
	engine := NewReflectionEngine(mockCC, "test-model")
	engine.SetThreshold(10.0)

	store := NewInMemoryStore()
	ms := NewMemoryStream("agent-r1", store, nil)
	ctx := context.Background()

	ms.Add(ctx, "nothing important", "observation", 0.1)

	insights, err := engine.MaybeReflect(ctx, ms, time.Now().Add(-1*time.Hour))
	require.NoError(t, err)
	assert.Nil(t, insights)
	assert.Empty(t, mockCC.Calls(), "should not call LLM when below threshold")
}

func TestReflectionEngine_AboveThreshold(t *testing.T) {
	mockCC := openclaw.NewMockChatClient()
	mockCC.PushResponse(*openclaw.SimpleTextResponse("1. What is the trend?\n2. Who are the key players?\n3. What are the risks?"))
	mockCC.PushResponse(*openclaw.SimpleTextResponse("The trend shows increasing polarization."))
	mockCC.PushResponse(*openclaw.SimpleTextResponse("The key players are government and media."))
	mockCC.PushResponse(*openclaw.SimpleTextResponse("The main risk is social unrest."))

	engine := NewReflectionEngine(mockCC, "test-model")
	engine.SetThreshold(2.0)

	store := NewInMemoryStore()
	ms := NewMemoryStream("agent-r2", store, nil)
	ctx := context.Background()

	ms.Add(ctx, "Major protest occurred", "observation", 0.9)
	ms.Add(ctx, "Government responded with force", "observation", 0.8)
	ms.Add(ctx, "Media coverage was intense", "observation", 0.7)

	past := time.Now().Add(-2 * time.Hour)
	insights, err := engine.MaybeReflect(ctx, ms, past)
	require.NoError(t, err)
	assert.Len(t, insights, 3)

	for _, insight := range insights {
		assert.Equal(t, "reflection", insight.Kind)
		assert.Equal(t, reflectionHighImportance, insight.Importance)
	}

	assert.Len(t, mockCC.Calls(), 4) // 1 question gen + 3 insight gen
}

func TestFormatMemories(t *testing.T) {
	entries := []struct {
		content string
		kind    string
		imp     float64
	}{
		{"event one", "observation", 0.5},
		{"event two", "action", 0.8},
	}
	_ = entries
	// Smoke test
	result := FormatMemories(nil)
	assert.Equal(t, "", result)
}
