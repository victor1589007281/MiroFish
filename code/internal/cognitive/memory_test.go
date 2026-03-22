package cognitive

import (
	"context"
	"testing"
	"time"

	"swarm-predict/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func simpleEmbedder(_ context.Context, text string) ([]float32, error) {
	v := make([]float32, 4)
	for i, ch := range text {
		if i >= 4 {
			break
		}
		v[i] = float32(ch) / 1000.0
	}
	return v, nil
}

func TestMemoryStream_AddAndRetrieve(t *testing.T) {
	store := NewInMemoryStore()
	ms := NewMemoryStream("agent-1", store, simpleEmbedder)
	ctx := context.Background()

	_, err := ms.Add(ctx, "I saw a protest in the square", "observation", 0.8)
	require.NoError(t, err)
	_, err = ms.Add(ctx, "The weather is nice today", "observation", 0.2)
	require.NoError(t, err)
	_, err = ms.Add(ctx, "The government announced a new policy", "observation", 0.9)
	require.NoError(t, err)

	results, err := ms.Retrieve(ctx, "political event", 2)
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "agent-1", results[0].AgentID)
}

func TestMemoryStream_EmptyRetrieval(t *testing.T) {
	store := NewInMemoryStore()
	ms := NewMemoryStream("agent-2", store, nil)
	ctx := context.Background()

	results, err := ms.Retrieve(ctx, "anything", 5)
	require.NoError(t, err)
	assert.Nil(t, results)
}

func TestMemoryStream_ImportanceWeight(t *testing.T) {
	store := NewInMemoryStore()
	ms := NewMemoryStream("agent-3", store, nil, WithWeights(0, 0, 1.0))
	ctx := context.Background()

	ms.Add(ctx, "low importance event", "observation", 0.1)
	ms.Add(ctx, "high importance event", "observation", 0.95)
	ms.Add(ctx, "medium importance event", "observation", 0.5)

	results, err := ms.Retrieve(ctx, "event", 3)
	require.NoError(t, err)
	require.Len(t, results, 3)
	assert.Equal(t, "high importance event", results[0].Content)
	assert.Equal(t, "medium importance event", results[1].Content)
	assert.Equal(t, "low importance event", results[2].Content)
}

func TestMemoryStream_RecencyWeight(t *testing.T) {
	store := NewInMemoryStore()
	ms := NewMemoryStream("agent-4", store, nil, WithWeights(1.0, 0, 0))
	ctx := context.Background()

	old := time.Now().Add(-48 * time.Hour)
	store.Append(ctx, model.MemoryEntry{
		ID:         "old-1",
		AgentID:    "agent-4",
		Content:    "very old event",
		Timestamp:  old,
		Importance: 0.5,
		Kind:       "observation",
	})
	ms.Add(ctx, "recent event", "observation", 0.5)

	results, err := ms.Retrieve(ctx, "event", 2)
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, "recent event", results[0].Content)
}

func TestExponentialDecay(t *testing.T) {
	assert.InDelta(t, 1.0, ExponentialDecay(0, 0.995), 0.001)
	decay1h := ExponentialDecay(time.Hour, 0.995)
	decay24h := ExponentialDecay(24*time.Hour, 0.995)
	assert.Greater(t, decay1h, decay24h)
	assert.Greater(t, decay1h, 0.0)
}

func TestCosineSimilarity(t *testing.T) {
	a := []float32{1, 0, 0}
	b := []float32{1, 0, 0}
	assert.InDelta(t, 1.0, CosineSimilarity(a, b), 0.001)

	c := []float32{0, 1, 0}
	assert.InDelta(t, 0.0, CosineSimilarity(a, c), 0.001)

	assert.Equal(t, 0.0, CosineSimilarity(nil, nil))
	assert.Equal(t, 0.0, CosineSimilarity([]float32{}, []float32{}))
}

func TestSumImportance(t *testing.T) {
	assert.InDelta(t, 0.0, SumImportance(nil), 0.001)
	entries := []model.MemoryEntry{
		{Importance: 0.3},
		{Importance: 0.7},
		{Importance: 1.0},
	}
	assert.InDelta(t, 2.0, SumImportance(entries), 0.001)
}

func TestInMemoryStore(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	now := time.Now()
	store.Append(ctx, model.MemoryEntry{ID: "1", AgentID: "a1", Content: "first", Timestamp: now.Add(-2 * time.Hour), Importance: 0.5, Kind: "observation"})
	store.Append(ctx, model.MemoryEntry{ID: "2", AgentID: "a1", Content: "second", Timestamp: now.Add(-1 * time.Hour), Importance: 0.8, Kind: "observation"})
	store.Append(ctx, model.MemoryEntry{ID: "3", AgentID: "a2", Content: "other agent", Timestamp: now, Importance: 0.3, Kind: "observation"})

	entries, err := store.GetRecent(ctx, "a1", 10)
	require.NoError(t, err)
	assert.Len(t, entries, 2)

	since, err := store.GetSince(ctx, "a1", now.Add(-90*time.Minute))
	require.NoError(t, err)
	assert.Len(t, since, 1)
	assert.Equal(t, "second", since[0].Content)

	all, err := store.GetAll(ctx, "a2")
	require.NoError(t, err)
	assert.Len(t, all, 1)
}
