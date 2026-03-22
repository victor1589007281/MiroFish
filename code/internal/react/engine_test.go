package react

import (
	"context"
	"testing"

	"swarm-predict/internal/model"
	"swarm-predict/internal/openclaw"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEngine_GenerateReport(t *testing.T) {
	mockCC := openclaw.NewMockChatClient()
	callIdx := 0
	mockCC.SetHandler(func(req openclaw.ChatRequest) (*openclaw.ChatResponse, error) {
		callIdx++
		if callIdx == 1 {
			return openclaw.SimpleTextResponse(`["Summary", "Analysis", "Predictions"]`), nil
		}
		return openclaw.SimpleTextResponse("This section covers the key findings from the simulation."), nil
	})

	engine := NewEngine(mockCC, "test-model")
	predictions := []model.PredictionResult{
		{Question: "Will X happen?", Probability: 0.75, Confidence: 0.85},
	}

	report, err := engine.GenerateReport(context.Background(), "Test project", "30 agents, 40 rounds", predictions)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(report.Sections), 3)
	assert.Len(t, report.Predictions, 1)

	for _, s := range report.Sections {
		assert.NotEmpty(t, s.Title)
		assert.NotEmpty(t, s.Content)
		assert.Greater(t, s.Order, 0)
	}
}

func TestEngine_Chat(t *testing.T) {
	mockCC := openclaw.NewMockChatClient()
	mockCC.PushResponse(*openclaw.SimpleTextResponse("The prediction shows 75% probability based on simulation data."))

	engine := NewEngine(mockCC, "test-model")
	report := &model.Report{
		Sections: []model.ReportSection{
			{Title: "Summary", Content: "This is a test report.", Order: 1},
		},
	}

	answer, err := engine.Chat(context.Background(), report, "What is the main prediction?")
	require.NoError(t, err)
	assert.Contains(t, answer, "75%")
}

func TestEngine_WithToolCalls(t *testing.T) {
	mockCC := openclaw.NewMockChatClient()

	// First call: plan outline
	mockCC.PushResponse(*openclaw.SimpleTextResponse(`["Analysis"]`))

	// Second call: LLM requests a tool call
	mockCC.PushResponse(*openclaw.ToolCallResponse("emergence_analysis", "{}"))

	// Third call: LLM produces final content after seeing tool result
	mockCC.PushResponse(*openclaw.SimpleTextResponse("Based on the emergence analysis, polarization is moderate."))

	mockTool := Tool{
		Definition: openclaw.ToolDef{
			Type: "function",
			Function: openclaw.ToolFuncDef{
				Name:        "emergence_analysis",
				Description: "Analyze emergence",
				Parameters:  map[string]interface{}{},
			},
		},
		Execute: func(_ context.Context, _ string) (string, error) {
			return `{"polarization_index": 0.35, "consensus_level": 0.65}`, nil
		},
	}

	engine := NewEngine(mockCC, "test-model", mockTool)
	report, err := engine.GenerateReport(context.Background(), "proj", "sim", nil)
	require.NoError(t, err)
	require.Len(t, report.Sections, 1)
	assert.Contains(t, report.Sections[0].Content, "polarization")
}
