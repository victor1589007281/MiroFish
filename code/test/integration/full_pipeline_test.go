//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"testing"

	"swarm-predict/internal/cognitive"
	"swarm-predict/internal/model"
	"swarm-predict/internal/openclaw"
	"swarm-predict/internal/prediction"
	"swarm-predict/internal/react"
	"swarm-predict/internal/simulation"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFullPipeline runs the complete MiroFish V2 pipeline with mock OpenClaw.
func TestFullPipeline(t *testing.T) {
	ctx := context.Background()

	// --- Setup mocks ---
	mockCC := openclaw.NewMockChatClient()
	mockSpawner := openclaw.NewMockSpawnClient()

	// Mock for ontology generation
	ontologyResp := model.OntologyResult{
		EntityTypes: []model.EntityType{
			{Name: "government", Description: "Government bodies"},
			{Name: "media", Description: "News media organizations"},
			{Name: "public", Description: "General public groups"},
		},
		RelationTypes: []model.RelationType{
			{Name: "influences", Source: "government", Target: "public"},
			{Name: "reports_on", Source: "media", Target: "government"},
		},
	}

	// Mock for GM arbitration and other CC calls
	mockCC.SetHandler(func(req openclaw.ChatRequest) (*openclaw.ChatResponse, error) {
		lastMsg := req.Messages[len(req.Messages)-1].Content

		// Ontology request
		if containsAny(lastMsg, "entity_types", "ontology", "Prediction goal") {
			return openclaw.JSONTextResponse(ontologyResp), nil
		}
		// GM arbitration
		if containsAny(lastMsg, "Game Master", "Agent:") {
			return openclaw.SimpleTextResponse(`{"approved": true, "reason": "consistent"}`), nil
		}
		// Reflection questions
		if containsAny(lastMsg, "generate 3 high-level questions", "analyzing") {
			return openclaw.SimpleTextResponse("1. What are the main trends?\n2. Who are the key influencers?\n3. What risks exist?"), nil
		}
		// Reflection insight
		if containsAny(lastMsg, "Question:", "Relevant memories") {
			return openclaw.SimpleTextResponse("The trend shows growing public concern about the policy changes."), nil
		}
		// Report outline
		if containsAny(lastMsg, "report outline", "section titles") {
			return openclaw.SimpleTextResponse(`["Executive Summary", "Key Findings", "Predictions"]`), nil
		}
		// Mediator
		if containsAny(lastMsg, "Expert predictions", "Analyze the expert") {
			return openclaw.SimpleTextResponse(`{"feedback": "experts roughly agree", "disagreements": ["timing uncertainty"], "key_arguments": ["data supports 70%"]}`), nil
		}

		return openclaw.SimpleTextResponse("This section presents our analysis of the simulation results."), nil
	})

	// Mock spawner for simulation and Delphi
	mockSpawner.SetHandler(func(req openclaw.SpawnRequest) (string, string) {
		if containsAny(req.Task, "sim-r", "Round") {
			// Simulation group decisions
			decisions := []map[string]string{
				{"agent_id": "gov-1", "action_type": "post", "content": "New policy announcement on economic reform"},
				{"agent_id": "media-1", "action_type": "post", "content": "Breaking: Government announces major reforms"},
				{"agent_id": "pub-1", "action_type": "reply", "content": "I support this change", "target_id": "post-1"},
			}
			data, _ := json.Marshal(decisions)
			return "run-" + req.Label, string(data)
		}
		// Delphi expert
		ep := prediction.ExpertOutput{
			Probability: 0.70,
			Reasoning:   "Simulation data and historical patterns support a 70% probability.",
		}
		data, _ := json.Marshal(ep)
		return "run-" + req.Label, string(data)
	})

	// --- Step 1: Ontology (via chatCompletions) ---
	t.Run("Step1_Ontology", func(t *testing.T) {
		resp, err := mockCC.Complete(ctx, openclaw.ChatRequest{
			Model: "test-model",
			Messages: []openclaw.ChatMessage{
				{Role: "system", Content: "Analyze entity_types"},
				{Role: "user", Content: "Document: test doc\nPrediction goal: test"},
			},
		})
		require.NoError(t, err)
		var onto model.OntologyResult
		err = cognitive.ExtractJSON(resp, &onto)
		require.NoError(t, err)
		assert.Len(t, onto.EntityTypes, 3)
		assert.Len(t, onto.RelationTypes, 2)
	})

	// --- Step 2: Agent profiles ---
	agents := []model.AgentProfile{
		{ID: "gov-1", Name: "Minister Zhang", Role: "government", Personality: "authoritative, reform-minded"},
		{ID: "media-1", Name: "Reporter Li", Role: "media", Personality: "investigative, skeptical"},
		{ID: "pub-1", Name: "Citizen Wang", Role: "public", Personality: "concerned, active on social media"},
		{ID: "pub-2", Name: "Student Chen", Role: "public", Personality: "young, tech-savvy, critical"},
		{ID: "analyst-1", Name: "Dr. Liu", Role: "analyst", Personality: "data-driven, neutral"},
		{ID: "biz-1", Name: "CEO Zhao", Role: "business", Personality: "pragmatic, profit-oriented"},
	}

	// --- Step 3: Simulation ---
	t.Run("Step3_Simulation", func(t *testing.T) {
		config := model.SimulationConfig{
			ID:             "integ-test-sim",
			Rounds:         5,
			AgentsPerGroup: 3,
			Model:          "test-model",
			FlashModel:     "test-flash",
			Events: []model.EventConfig{
				{Round: 2, Content: "Breaking: Economic data shows unexpected downturn"},
			},
		}

		engine := simulation.NewEngine(mockSpawner, mockCC, config)
		result, err := engine.Run(ctx, agents)
		require.NoError(t, err)
		assert.Equal(t, 5, result.TotalRounds)
		assert.Greater(t, result.TotalPosts, 0)
		assert.Greater(t, len(result.ActionLogs), 0)

		// Emergence detection
		signal := simulation.DetectEmergence(result.ActionLogs)
		t.Logf("Emergence: polarization=%.3f, consensus=%.3f, topics=%v",
			signal.PolarizationIndex, signal.ConsensusLevel, signal.TopTopics)
	})

	// --- Step 4: Prediction Refinement ---
	t.Run("Step4_PredictionRefinement", func(t *testing.T) {
		delphi := prediction.NewDelphiEngine(mockSpawner, mockCC, "test-model")
		delphi.SetMaxRounds(2)

		results, err := delphi.Refine(ctx, "30 agents simulated 5 rounds of social interaction", []string{
			"Will the economic reform policy be passed within 6 months?",
			"Will public approval rating exceed 60% after implementation?",
		})
		require.NoError(t, err)
		assert.Len(t, results, 2)

		for _, r := range results {
			assert.Greater(t, r.Probability, 0.01)
			assert.Less(t, r.Probability, 0.99)
			assert.Greater(t, r.Confidence, 0.0)
			assert.Greater(t, len(r.ExpertViews), 0)
			t.Logf("Q: %s → P=%.2f, C=%.2f, Rounds=%d", r.Question, r.Probability, r.Confidence, r.Rounds)
		}
	})

	// --- Step 5: Report Generation ---
	t.Run("Step5_ReportGeneration", func(t *testing.T) {
		predictions := []model.PredictionResult{
			{Question: "Reform passes?", Probability: 0.72, Confidence: 0.85},
			{Question: "Approval > 60%?", Probability: 0.55, Confidence: 0.70},
		}

		engine := react.NewEngine(mockCC, "test-model")
		report, err := engine.GenerateReport(ctx, "Economic reform prediction", "6 agents, 5 rounds", predictions)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(report.Sections), 3)

		for _, s := range report.Sections {
			assert.NotEmpty(t, s.Title)
			assert.NotEmpty(t, s.Content)
			t.Logf("Section: %s (%d chars)", s.Title, len(s.Content))
		}
	})

	// --- Step 6: Chat ---
	t.Run("Step6_Chat", func(t *testing.T) {
		report := &model.Report{
			Sections: []model.ReportSection{
				{Title: "Summary", Content: "The reform has 72% probability of passing.", Order: 1},
			},
		}
		engine := react.NewEngine(mockCC, "test-model")
		answer, err := engine.Chat(ctx, report, "What is the main risk?")
		require.NoError(t, err)
		assert.NotEmpty(t, answer)
	})

	// --- Cognitive Engine: Memory + Reflection ---
	t.Run("CognitiveEngine_MemoryAndReflection", func(t *testing.T) {
		store := cognitive.NewInMemoryStore()
		ms := cognitive.NewMemoryStream("agent-test", store, nil)

		ms.Add(ctx, "I observed a heated debate about the reform", "observation", 0.8)
		ms.Add(ctx, "Minister Zhang made a strong speech", "observation", 0.9)
		ms.Add(ctx, "Public sentiment seems divided", "observation", 0.7)
		ms.Add(ctx, "Media coverage is intensifying", "observation", 0.6)

		memories, err := ms.Retrieve(ctx, "reform debate", 3)
		require.NoError(t, err)
		assert.Len(t, memories, 3)

		reflEngine := cognitive.NewReflectionEngine(mockCC, "test-model")
		reflEngine.SetThreshold(2.0)

		insights, err := reflEngine.MaybeReflect(ctx, ms, memories[0].Timestamp.Add(-1))
		require.NoError(t, err)
		assert.NotNil(t, insights)
		assert.Greater(t, len(insights), 0)
		t.Logf("Generated %d reflection insights", len(insights))
	})

	// --- Social Graph ---
	t.Run("SocialGraph", func(t *testing.T) {
		socialStore := cognitive.NewInMemorySocialStore()
		sg := cognitive.NewSocialGraph(socialStore)

		sg.UpdateAfterInteraction(ctx, "pub-1", "gov-1", cognitive.InteractionDelta{
			LikabilityDelta: 0.3,
			ReputationDelta: 0.2,
		})
		sg.UpdateAfterInteraction(ctx, "pub-1", "media-1", cognitive.InteractionDelta{
			LikabilityDelta: -0.1,
			ReputationDelta: 0.4,
		})

		rels, err := sg.GetRelationships(ctx, "pub-1")
		require.NoError(t, err)
		assert.Len(t, rels, 2)
	})
}

func containsAny(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if len(s) > 0 && len(sub) > 0 {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
		}
	}
	return false
}

// TestCalibrationPipeline tests the full calibration math.
func TestCalibrationPipeline(t *testing.T) {
	cal := prediction.NewCalibrator()

	testCases := []struct {
		name     string
		rawProb  float64
		wantMin  float64
		wantMax  float64
	}{
		{"very_high", 0.95, 0.5, 0.99},
		{"high", 0.80, 0.5, 0.99},
		{"medium", 0.50, 0.01, 0.99},
		{"low", 0.20, 0.01, 0.5},
		{"very_low", 0.05, 0.01, 0.5},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := cal.Calibrate(tc.rawProb)
			assert.GreaterOrEqual(t, result, tc.wantMin, "probability should be >= %f", tc.wantMin)
			assert.LessOrEqual(t, result, tc.wantMax, "probability should be <= %f", tc.wantMax)
			t.Logf("%.2f -> %.4f", tc.rawProb, result)
		})
	}
}

// TestDelphiConvergence tests that the Delphi process converges correctly.
func TestDelphiConvergence(t *testing.T) {
	mockSpawner := openclaw.NewMockSpawnClient()
	mockCC := openclaw.NewMockChatClient()

	var roundCounter int64
	mockSpawner.SetHandler(func(req openclaw.SpawnRequest) (string, string) {
		r := atomic.AddInt64(&roundCounter, 1)
		prob := 0.5 + float64(r)*0.02
		ep := prediction.ExpertOutput{
			Probability: prob,
			Reasoning:   fmt.Sprintf("round %d analysis", r),
		}
		data, _ := json.Marshal(ep)
		return fmt.Sprintf("run-%d-%s", r, req.Label), string(data)
	})
	mockCC.SetHandler(func(req openclaw.ChatRequest) (*openclaw.ChatResponse, error) {
		return openclaw.SimpleTextResponse(`{"feedback": "align more", "disagreements": [], "key_arguments": []}`), nil
	})

	delphi := prediction.NewDelphiEngine(mockSpawner, mockCC, "m")
	delphi.SetMaxRounds(3)

	results, err := delphi.Refine(context.Background(), "test", []string{"Q?"})
	require.NoError(t, err)
	require.Len(t, results, 1)

	r := results[0]
	assert.Greater(t, r.Probability, 0.0)
	assert.Greater(t, r.Rounds, 0)
	assert.LessOrEqual(t, r.Rounds, 3)
}
