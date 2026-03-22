package agent

import (
	"context"
	"testing"

	"swarm-predict/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockBlock struct {
	name     string
	active   bool
	output   *BlockOutput
}

func (b *mockBlock) Name() string                                          { return b.name }
func (b *mockBlock) ShouldActivate(_ *StepContext) bool                    { return b.active }
func (b *mockBlock) Execute(_ context.Context, _ *StepContext) (*BlockOutput, error) { return b.output, nil }

func TestDispatcher_Dispatch(t *testing.T) {
	memBlock := &mockBlock{name: "memory", active: true, output: &BlockOutput{
		BlockName: "memory", Intent: "loaded", Details: map[string]string{"retrieved_memories": "some"},
	}}
	postBlock := &mockBlock{name: "posting", active: true, output: &BlockOutput{
		BlockName: "posting", Intent: "wants to post", Details: map[string]string{
			"action_type": "post", "content": "hello world",
		},
	}}

	ag := NewAgent("a1", model.AgentProfile{ID: "a1", Name: "Test"}, nil, nil, memBlock, postBlock)
	dispatcher := NewDispatcher()

	proposal, err := dispatcher.Dispatch(context.Background(), ag, &StepContext{Round: 1})
	require.NoError(t, err)
	assert.Equal(t, "a1", proposal.AgentID)
	assert.Equal(t, "post", proposal.ActionType)
	assert.Equal(t, "hello world", proposal.Content)
}

func TestDispatcher_NoActionableOutput(t *testing.T) {
	memBlock := &mockBlock{name: "memory", active: true, output: &BlockOutput{
		BlockName: "memory", Intent: "loaded", Details: map[string]string{},
	}}

	ag := NewAgent("a2", model.AgentProfile{ID: "a2"}, nil, nil, memBlock)
	dispatcher := NewDispatcher()

	proposal, err := dispatcher.Dispatch(context.Background(), ag, &StepContext{Round: 1})
	require.NoError(t, err)
	assert.Equal(t, "like", proposal.ActionType) // fallback
}

func TestDispatcher_PriorityResolution(t *testing.T) {
	reactionBlock := &mockBlock{name: "reaction", active: true, output: &BlockOutput{
		BlockName: "reaction", Intent: "react", Details: map[string]string{
			"action_type": "reply", "content": "I agree", "target_id": "p1",
		},
	}}
	postBlock := &mockBlock{name: "posting", active: true, output: &BlockOutput{
		BlockName: "posting", Intent: "wants to post", Details: map[string]string{
			"action_type": "post", "content": "new post",
		},
	}}

	ag := NewAgent("a3", model.AgentProfile{ID: "a3"}, nil, nil, reactionBlock, postBlock)
	dispatcher := NewDispatcher()

	proposal, err := dispatcher.Dispatch(context.Background(), ag, &StepContext{Round: 1})
	require.NoError(t, err)
	assert.Equal(t, "post", proposal.ActionType) // posting has higher priority
}

func TestDispatcher_DispatchAll(t *testing.T) {
	block := &mockBlock{name: "posting", active: true, output: &BlockOutput{
		BlockName: "posting", Intent: "post", Details: map[string]string{"action_type": "post", "content": "hello"},
	}}

	agents := []*Agent{
		NewAgent("a1", model.AgentProfile{ID: "a1"}, nil, nil, block),
		NewAgent("a2", model.AgentProfile{ID: "a2"}, nil, nil, block),
	}

	dispatcher := NewDispatcher()
	proposals, err := dispatcher.DispatchAll(context.Background(), agents, &StepContext{Round: 1})
	require.NoError(t, err)
	assert.Len(t, proposals, 2)
}
