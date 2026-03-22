package cognitive

import (
	"context"
	"math"
	"sync"
	"time"

	"swarm-predict/internal/model"
)

// SocialGraphStore persists social relations.
type SocialGraphStore interface {
	Get(ctx context.Context, observerID, targetID string) (*model.SocialRelation, error)
	Set(ctx context.Context, rel model.SocialRelation) error
	GetRelationships(ctx context.Context, agentID string) ([]model.SocialRelation, error)
	GetInfluencers(ctx context.Context, topK int) ([]model.SocialRelation, error)
}

// SocialGraph manages inter-agent perception (reputation, influence, likability).
type SocialGraph struct {
	store SocialGraphStore
}

func NewSocialGraph(store SocialGraphStore) *SocialGraph {
	return &SocialGraph{store: store}
}

// InteractionDelta describes how an interaction changes social perception.
type InteractionDelta struct {
	LikabilityDelta float64
	ReputationDelta float64
	InfluenceDelta  float64
}

func (sg *SocialGraph) UpdateAfterInteraction(ctx context.Context, observerID, targetID string, delta InteractionDelta) error {
	rel, err := sg.store.Get(ctx, observerID, targetID)
	if err != nil {
		rel = &model.SocialRelation{
			ObserverID: observerID,
			TargetID:   targetID,
		}
	}

	rel.Likability = clamp(rel.Likability+delta.LikabilityDelta, -1, 1)
	rel.Reputation = clamp(rel.Reputation+delta.ReputationDelta, -1, 1)
	rel.Influence = clamp(rel.Influence+delta.InfluenceDelta, 0, 1)
	rel.LastUpdated = time.Now()

	return sg.store.Set(ctx, *rel)
}

func (sg *SocialGraph) GetRelationships(ctx context.Context, agentID string) ([]model.SocialRelation, error) {
	return sg.store.GetRelationships(ctx, agentID)
}

func (sg *SocialGraph) GetInfluencers(ctx context.Context, topK int) ([]model.SocialRelation, error) {
	return sg.store.GetInfluencers(ctx, topK)
}

func clamp(v, min, max float64) float64 {
	return math.Max(min, math.Min(max, v))
}

// InMemorySocialStore is a test-friendly in-memory implementation.
type InMemorySocialStore struct {
	mu   sync.Mutex
	rels map[string]model.SocialRelation // "observer:target" -> relation
}

func NewInMemorySocialStore() *InMemorySocialStore {
	return &InMemorySocialStore{rels: make(map[string]model.SocialRelation)}
}

func key(a, b string) string { return a + ":" + b }

func (s *InMemorySocialStore) Get(_ context.Context, observerID, targetID string) (*model.SocialRelation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rel, ok := s.rels[key(observerID, targetID)]
	if !ok {
		return &model.SocialRelation{ObserverID: observerID, TargetID: targetID}, nil
	}
	return &rel, nil
}

func (s *InMemorySocialStore) Set(_ context.Context, rel model.SocialRelation) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rels[key(rel.ObserverID, rel.TargetID)] = rel
	return nil
}

func (s *InMemorySocialStore) GetRelationships(_ context.Context, agentID string) ([]model.SocialRelation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var result []model.SocialRelation
	for _, rel := range s.rels {
		if rel.ObserverID == agentID {
			result = append(result, rel)
		}
	}
	return result, nil
}

func (s *InMemorySocialStore) GetInfluencers(_ context.Context, topK int) ([]model.SocialRelation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var all []model.SocialRelation
	for _, rel := range s.rels {
		all = append(all, rel)
	}
	if len(all) <= topK {
		return all, nil
	}
	return all[:topK], nil
}
