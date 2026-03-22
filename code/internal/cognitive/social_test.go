package cognitive

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSocialGraph_UpdateAndGet(t *testing.T) {
	store := NewInMemorySocialStore()
	sg := NewSocialGraph(store)
	ctx := context.Background()

	err := sg.UpdateAfterInteraction(ctx, "alice", "bob", InteractionDelta{
		LikabilityDelta: 0.3,
		ReputationDelta: 0.2,
		InfluenceDelta:  0.1,
	})
	require.NoError(t, err)

	rels, err := sg.GetRelationships(ctx, "alice")
	require.NoError(t, err)
	require.Len(t, rels, 1)
	assert.Equal(t, "bob", rels[0].TargetID)
	assert.InDelta(t, 0.3, rels[0].Likability, 0.01)
	assert.InDelta(t, 0.2, rels[0].Reputation, 0.01)
}

func TestSocialGraph_ClampValues(t *testing.T) {
	store := NewInMemorySocialStore()
	sg := NewSocialGraph(store)
	ctx := context.Background()

	for i := 0; i < 10; i++ {
		sg.UpdateAfterInteraction(ctx, "a", "b", InteractionDelta{
			LikabilityDelta: 0.5,
			ReputationDelta: 0.5,
			InfluenceDelta:  0.5,
		})
	}

	rels, _ := sg.GetRelationships(ctx, "a")
	require.Len(t, rels, 1)
	assert.LessOrEqual(t, rels[0].Likability, 1.0)
	assert.LessOrEqual(t, rels[0].Reputation, 1.0)
	assert.LessOrEqual(t, rels[0].Influence, 1.0)
}

func TestSocialGraph_NegativeInteraction(t *testing.T) {
	store := NewInMemorySocialStore()
	sg := NewSocialGraph(store)
	ctx := context.Background()

	sg.UpdateAfterInteraction(ctx, "x", "y", InteractionDelta{LikabilityDelta: 0.5})
	sg.UpdateAfterInteraction(ctx, "x", "y", InteractionDelta{LikabilityDelta: -0.8})

	rels, _ := sg.GetRelationships(ctx, "x")
	require.Len(t, rels, 1)
	assert.InDelta(t, -0.3, rels[0].Likability, 0.01)
	assert.GreaterOrEqual(t, rels[0].Likability, -1.0)
}

func TestClamp(t *testing.T) {
	assert.Equal(t, 0.5, clamp(0.5, 0, 1))
	assert.Equal(t, 0.0, clamp(-0.1, 0, 1))
	assert.Equal(t, 1.0, clamp(1.5, 0, 1))
	assert.Equal(t, -1.0, clamp(-2.0, -1, 1))
}
