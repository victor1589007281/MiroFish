package prediction

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"swarm-predict/internal/openclaw"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDelphiEngine_Refine(t *testing.T) {
	mockSpawner := openclaw.NewMockSpawnClient()
	mockCC := openclaw.NewMockChatClient()

	// Mock experts: return slightly different probabilities based on label
	mockSpawner.SetHandler(func(req openclaw.SpawnRequest) (string, string) {
		probMap := map[string]float64{
			"optimist": 0.72, "pessimist": 0.60, "quant": 0.68,
			"domain_expert": 0.65, "strategist": 0.70,
		}
		prob := 0.66
		for key, p := range probMap {
			if strings.Contains(req.Label, key) {
				prob = p
				break
			}
		}
		ep := ExpertOutput{
			Probability: prob,
			Reasoning:   "test reasoning for " + req.Label,
		}
		data, _ := json.Marshal(ep)
		return "run-" + req.Label, string(data)
	})

	// Mock mediator response
	mockCC.SetHandler(func(req openclaw.ChatRequest) (*openclaw.ChatResponse, error) {
		return openclaw.SimpleTextResponse(`{"feedback": "experts are fairly aligned", "disagreements": ["minor spread"], "key_arguments": ["data supports ~0.67"]}`), nil
	})

	engine := NewDelphiEngine(mockSpawner, mockCC, "test-model")
	engine.SetMaxRounds(2)

	results, err := engine.Refine(context.Background(), "simulation showed positive trend", []string{"Will X happen by 2027?"})
	require.NoError(t, err)
	require.Len(t, results, 1)

	r := results[0]
	assert.Equal(t, "Will X happen by 2027?", r.Question)
	assert.Greater(t, r.Probability, 0.01)
	assert.Less(t, r.Probability, 0.99)
	assert.Greater(t, r.Confidence, 0.0)
	assert.Greater(t, len(r.ExpertViews), 0)
	assert.Greater(t, r.Rounds, 0)
}

func TestDelphiEngine_ConvergesEarly(t *testing.T) {
	mockSpawner := openclaw.NewMockSpawnClient()
	mockCC := openclaw.NewMockChatClient()

	// All experts return same probability -> should converge in round 1
	mockSpawner.SetHandler(func(req openclaw.SpawnRequest) (string, string) {
		ep := ExpertOutput{Probability: 0.75, Reasoning: "all agree"}
		data, _ := json.Marshal(ep)
		return "run-" + req.Label, string(data)
	})

	engine := NewDelphiEngine(mockSpawner, mockCC, "test-model")
	engine.SetMaxRounds(5)

	results, err := engine.Refine(context.Background(), "summary", []string{"Q?"})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, 1, results[0].Rounds)
}

func TestDelphiEngine_MultipleQuestions(t *testing.T) {
	mockSpawner := openclaw.NewMockSpawnClient()
	mockCC := openclaw.NewMockChatClient()

	mockSpawner.SetHandler(func(req openclaw.SpawnRequest) (string, string) {
		prob := 0.5
		if strings.Contains(req.Task, "Q1") {
			prob = 0.8
		} else {
			prob = 0.3
		}
		ep := ExpertOutput{Probability: prob, Reasoning: "reasoning"}
		data, _ := json.Marshal(ep)
		return "run-" + req.Label, string(data)
	})

	mockCC.SetHandler(func(req openclaw.ChatRequest) (*openclaw.ChatResponse, error) {
		return openclaw.SimpleTextResponse(`{"feedback": "ok", "disagreements": [], "key_arguments": []}`), nil
	})

	engine := NewDelphiEngine(mockSpawner, mockCC, "m")
	engine.SetMaxRounds(1)

	results, err := engine.Refine(context.Background(), "data", []string{"Q1: positive?", "Q2: negative?"})
	require.NoError(t, err)
	assert.Len(t, results, 2)
}
