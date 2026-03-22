package simulation

import (
	"context"
	"encoding/json"
	"testing"

	"swarm-predict/internal/model"
	"swarm-predict/internal/openclaw"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPersonaGenerator_Generate(t *testing.T) {
	mockCC := openclaw.NewMockChatClient()

	profiles := []model.AgentProfile{
		{ID: "a1", Name: "Zhang Wei", Role: "government", Personality: "authoritative"},
		{ID: "a2", Name: "Li Na", Role: "media", Personality: "investigative"},
		{ID: "a3", Name: "Wang Min", Role: "public", Personality: "concerned"},
	}
	data, _ := json.Marshal(profiles)
	mockCC.PushResponse(*openclaw.SimpleTextResponse(string(data)))

	gen := NewPersonaGenerator(mockCC, "test-model")
	result, err := gen.Generate(context.Background(), "economic reform", 3)
	require.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Equal(t, "Zhang Wei", result[0].Name)
	assert.Equal(t, "government", result[0].Role)
}

func TestPersonaGenerator_EnsureIDs(t *testing.T) {
	mockCC := openclaw.NewMockChatClient()

	profiles := []model.AgentProfile{
		{Name: "NoID Agent", Role: "test"},
	}
	data, _ := json.Marshal(profiles)
	mockCC.PushResponse(*openclaw.SimpleTextResponse(string(data)))

	gen := NewPersonaGenerator(mockCC, "m")
	result, err := gen.Generate(context.Background(), "topic", 1)
	require.NoError(t, err)
	assert.Equal(t, "agent-1", result[0].ID) // auto-assigned
}
