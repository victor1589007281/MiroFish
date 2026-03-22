package cognitive

import (
	"context"
	"math"
	"sort"
	"sync"
	"time"

	"swarm-predict/internal/model"

	"github.com/google/uuid"
)

const (
	defaultDecayRate    = 0.995
	defaultWRecency     = 1.0
	defaultWRelevance   = 1.0
	defaultWImportance  = 1.0
	defaultMaxCandidates = 200
)

// EmbeddingFunc produces a vector from text. Inject implementation at runtime.
type EmbeddingFunc func(ctx context.Context, text string) ([]float32, error)

// MemoryStore is the persistence backend for memory entries.
type MemoryStore interface {
	Append(ctx context.Context, entry model.MemoryEntry) error
	GetRecent(ctx context.Context, agentID string, limit int) ([]model.MemoryEntry, error)
	GetSince(ctx context.Context, agentID string, since time.Time) ([]model.MemoryEntry, error)
	GetAll(ctx context.Context, agentID string) ([]model.MemoryEntry, error)
}

// MemoryStream manages an agent's memory with weighted retrieval.
type MemoryStream struct {
	agentID      string
	store        MemoryStore
	embedder     EmbeddingFunc
	wRecency     float64
	wRelevance   float64
	wImportance  float64
	decayRate    float64
}

type MemoryStreamOption func(*MemoryStream)

func WithWeights(recency, relevance, importance float64) MemoryStreamOption {
	return func(ms *MemoryStream) {
		ms.wRecency = recency
		ms.wRelevance = relevance
		ms.wImportance = importance
	}
}

func WithDecayRate(rate float64) MemoryStreamOption {
	return func(ms *MemoryStream) { ms.decayRate = rate }
}

func NewMemoryStream(agentID string, store MemoryStore, embedder EmbeddingFunc, opts ...MemoryStreamOption) *MemoryStream {
	ms := &MemoryStream{
		agentID:     agentID,
		store:       store,
		embedder:    embedder,
		wRecency:    defaultWRecency,
		wRelevance:  defaultWRelevance,
		wImportance: defaultWImportance,
		decayRate:   defaultDecayRate,
	}
	for _, o := range opts {
		o(ms)
	}
	return ms
}

// Add stores a new memory entry, computing embedding if available.
func (ms *MemoryStream) Add(ctx context.Context, content string, kind string, importance float64) (*model.MemoryEntry, error) {
	entry := model.MemoryEntry{
		ID:         uuid.New().String(),
		AgentID:    ms.agentID,
		Content:    content,
		Timestamp:  time.Now(),
		Importance: importance,
		Kind:       kind,
	}
	if ms.embedder != nil {
		emb, err := ms.embedder(ctx, content)
		if err == nil {
			entry.Embedding = emb
		}
	}
	if err := ms.store.Append(ctx, entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

// Retrieve returns top-K memories scored by recency × importance × relevance.
func (ms *MemoryStream) Retrieve(ctx context.Context, query string, k int) ([]model.MemoryEntry, error) {
	candidates, err := ms.store.GetRecent(ctx, ms.agentID, defaultMaxCandidates)
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return nil, nil
	}

	var queryEmb []float32
	if ms.embedder != nil {
		queryEmb, _ = ms.embedder(ctx, query)
	}

	now := time.Now()
	type scored struct {
		entry model.MemoryEntry
		score float64
	}
	scoredList := make([]scored, 0, len(candidates))

	for _, entry := range candidates {
		recency := ExponentialDecay(now.Sub(entry.Timestamp), ms.decayRate)
		relevance := 0.5 // default if no embedding
		if queryEmb != nil && entry.Embedding != nil {
			relevance = CosineSimilarity(queryEmb, entry.Embedding)
		}
		importance := entry.Importance

		total := ms.wRecency*recency + ms.wRelevance*relevance + ms.wImportance*importance
		scoredList = append(scoredList, scored{entry: entry, score: total})
	}

	sort.Slice(scoredList, func(i, j int) bool {
		return scoredList[i].score > scoredList[j].score
	})

	if k > len(scoredList) {
		k = len(scoredList)
	}
	result := make([]model.MemoryEntry, k)
	for i := 0; i < k; i++ {
		result[i] = scoredList[i].entry
	}
	return result, nil
}

// GetSince returns all memories since a given time.
func (ms *MemoryStream) GetSince(ctx context.Context, since time.Time) ([]model.MemoryEntry, error) {
	return ms.store.GetSince(ctx, ms.agentID, since)
}

// SumImportance calculates total importance for a set of entries.
func SumImportance(entries []model.MemoryEntry) float64 {
	var total float64
	for _, e := range entries {
		total += e.Importance
	}
	return total
}

// ExponentialDecay computes decay factor for a duration.
func ExponentialDecay(elapsed time.Duration, rate float64) float64 {
	hours := elapsed.Hours()
	return math.Pow(rate, hours)
}

// CosineSimilarity between two vectors.
func CosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}
	denom := math.Sqrt(normA) * math.Sqrt(normB)
	if denom == 0 {
		return 0
	}
	return dot / denom
}

// InMemoryStore is a simple in-process MemoryStore for testing.
type InMemoryStore struct {
	mu      sync.Mutex
	entries map[string][]model.MemoryEntry // agentID -> entries
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{entries: make(map[string][]model.MemoryEntry)}
}

func (s *InMemoryStore) Append(_ context.Context, entry model.MemoryEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[entry.AgentID] = append(s.entries[entry.AgentID], entry)
	return nil
}

func (s *InMemoryStore) GetRecent(_ context.Context, agentID string, limit int) ([]model.MemoryEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	all := s.entries[agentID]
	if len(all) <= limit {
		return append([]model.MemoryEntry{}, all...), nil
	}
	return append([]model.MemoryEntry{}, all[len(all)-limit:]...), nil
}

func (s *InMemoryStore) GetSince(_ context.Context, agentID string, since time.Time) ([]model.MemoryEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var result []model.MemoryEntry
	for _, e := range s.entries[agentID] {
		if !e.Timestamp.Before(since) {
			result = append(result, e)
		}
	}
	return result, nil
}

func (s *InMemoryStore) GetAll(_ context.Context, agentID string) ([]model.MemoryEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]model.MemoryEntry{}, s.entries[agentID]...), nil
}
