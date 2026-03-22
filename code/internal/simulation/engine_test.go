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

func TestEngine_RunSimulation(t *testing.T) {
	mockSpawner := openclaw.NewMockSpawnClient()
	mockCC := openclaw.NewMockChatClient()

	// Mock spawner: return valid group decisions
	mockSpawner.SetHandler(func(req openclaw.SpawnRequest) (string, string) {
		decisions := []map[string]string{
			{"agent_id": "a1", "action_type": "post", "content": "This is interesting"},
			{"agent_id": "a2", "action_type": "like", "target_id": "post-1"},
		}
		data, _ := json.Marshal(decisions)
		return "run-" + req.Label, string(data)
	})

	// Mock CC for GM arbitration (auto-approve posts)
	mockCC.SetHandler(func(req openclaw.ChatRequest) (*openclaw.ChatResponse, error) {
		return openclaw.SimpleTextResponse(`{"approved": true, "reason": "consistent with persona"}`), nil
	})

	agents := []model.AgentProfile{
		{ID: "a1", Name: "Alice", Role: "journalist", Personality: "curious"},
		{ID: "a2", Name: "Bob", Role: "student", Personality: "skeptical"},
		{ID: "a3", Name: "Carol", Role: "analyst", Personality: "data-driven"},
	}

	config := model.SimulationConfig{
		ID:             "test-sim",
		Rounds:         3,
		AgentsPerGroup: 3,
		Model:          "test-model",
		FlashModel:     "test-flash",
	}

	engine := NewEngine(mockSpawner, mockCC, config)
	result, err := engine.Run(context.Background(), agents)
	require.NoError(t, err)
	assert.Equal(t, 3, result.TotalRounds)
	assert.Greater(t, result.TotalPosts, 0)
	assert.Greater(t, len(result.ActionLogs), 0)
}

func TestEngine_EventInjection(t *testing.T) {
	mockSpawner := openclaw.NewMockSpawnClient()
	mockCC := openclaw.NewMockChatClient()
	mockCC.SetHandler(func(req openclaw.ChatRequest) (*openclaw.ChatResponse, error) {
		return openclaw.SimpleTextResponse(`{"approved": true, "reason": "ok"}`), nil
	})
	mockSpawner.SetHandler(func(req openclaw.SpawnRequest) (string, string) {
		return "run-evt", `[{"agent_id": "a1", "action_type": "post", "content": "reacting to event"}]`
	})

	config := model.SimulationConfig{
		Rounds:         2,
		AgentsPerGroup: 1,
		Model:          "m",
		FlashModel:     "m",
		Events: []model.EventConfig{
			{Round: 1, Content: "Breaking: major policy change announced"},
		},
	}

	engine := NewEngine(mockSpawner, mockCC, config)
	agents := []model.AgentProfile{{ID: "a1", Name: "A", Role: "r"}}
	result, err := engine.Run(context.Background(), agents)
	require.NoError(t, err)
	assert.Greater(t, result.TotalPosts, 1) // event post + agent posts
}

func TestGroupAgents(t *testing.T) {
	agents := make([]model.AgentProfile, 13)
	for i := range agents {
		agents[i] = model.AgentProfile{ID: "a" + string(rune('0'+i))}
	}

	groups := GroupAgents(agents, 5)
	assert.Len(t, groups, 3)
	assert.Len(t, groups[0], 5)
	assert.Len(t, groups[1], 5)
	assert.Len(t, groups[2], 3)
}

func TestGroupAgents_Empty(t *testing.T) {
	groups := GroupAgents(nil, 5)
	assert.Nil(t, groups)
}

func TestPlatform_CRUD(t *testing.T) {
	p := NewPlatform()
	post := p.CreatePost("user1", "hello world")
	assert.NotEmpty(t, post.ID)

	p.Like(post.ID)
	p.Like(post.ID)
	p.Reply(post.ID, "user2", "nice post")
	p.Repost(post.ID)

	feed := p.GetFeed(10)
	require.Len(t, feed, 1)
	assert.Equal(t, 2, feed[0].Likes)
	assert.Equal(t, 1, feed[0].Reposts)
	assert.Len(t, feed[0].Replies, 1)
}

func TestGameMaster_AutoApproveSimple(t *testing.T) {
	mockCC := openclaw.NewMockChatClient()
	gm := NewGameMaster(mockCC, "test")

	verdict, err := gm.Arbitrate(context.Background(),
		model.ActionProposal{AgentID: "a1", ActionType: "like", TargetID: "p1"},
		model.AgentProfile{}, nil)

	require.NoError(t, err)
	assert.True(t, verdict.Approved)
	assert.Empty(t, mockCC.Calls(), "should not call LLM for simple actions")
}

func TestGameMaster_LLMArbitratePost(t *testing.T) {
	mockCC := openclaw.NewMockChatClient()
	mockCC.PushResponse(*openclaw.SimpleTextResponse(`{"approved": false, "reason": "out of character", "suggestion": "tone it down"}`))

	gm := NewGameMaster(mockCC, "test")
	verdict, err := gm.Arbitrate(context.Background(),
		model.ActionProposal{AgentID: "a1", ActionType: "post", Content: "radical content"},
		model.AgentProfile{Name: "Conservative Bob", Personality: "reserved"}, nil)

	require.NoError(t, err)
	assert.False(t, verdict.Approved)
	assert.Contains(t, verdict.Reason, "out of character")
	assert.NotNil(t, verdict.ModifiedAction)
	assert.Len(t, mockCC.Calls(), 1)
}

func TestEmergenceDetection(t *testing.T) {
	logs := []ActionLog{
		{Round: 1, Decision: model.Decision{AgentID: "a1", ActionType: "post", Content: "This policy reform is great and positive for economy"}},
		{Round: 1, Decision: model.Decision{AgentID: "a2", ActionType: "post", Content: "This policy reform is terrible and negative for economy"}},
		{Round: 1, Decision: model.Decision{AgentID: "a1", ActionType: "post", Content: "Good support for the reform policy direction"}},
		{Round: 1, Decision: model.Decision{AgentID: "a2", ActionType: "post", Content: "I disagree with reform policy completely wrong"}},
	}

	signal := DetectEmergence(logs)
	assert.GreaterOrEqual(t, signal.PolarizationIndex, 0.0)
	assert.GreaterOrEqual(t, len(signal.TopTopics), 0)
	assert.GreaterOrEqual(t, signal.ConsensusLevel, 0.0)
	assert.LessOrEqual(t, signal.ConsensusLevel, 1.0)
}
