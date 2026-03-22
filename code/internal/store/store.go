package store

import (
	"context"
	"sync"

	"swarm-predict/internal/model"
)

// ProjectStore manages project persistence.
type ProjectStore interface {
	SaveSimulationState(ctx context.Context, state model.SimulationState) error
	GetSimulationState(ctx context.Context, id string) (*model.SimulationState, error)
	SaveReport(ctx context.Context, report model.Report) error
	GetReport(ctx context.Context, id string) (*model.Report, error)
}

// InMemoryStore is a test-friendly store implementation.
type InMemoryStore struct {
	mu          sync.RWMutex
	simStates   map[string]model.SimulationState
	reports     map[string]model.Report
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		simStates: make(map[string]model.SimulationState),
		reports:   make(map[string]model.Report),
	}
}

func (s *InMemoryStore) SaveSimulationState(_ context.Context, state model.SimulationState) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.simStates[state.ID] = state
	return nil
}

func (s *InMemoryStore) GetSimulationState(_ context.Context, id string) (*model.SimulationState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	state, ok := s.simStates[id]
	if !ok {
		return nil, nil
	}
	return &state, nil
}

func (s *InMemoryStore) SaveReport(_ context.Context, report model.Report) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reports[report.ID] = report
	return nil
}

func (s *InMemoryStore) GetReport(_ context.Context, id string) (*model.Report, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	report, ok := s.reports[id]
	if !ok {
		return nil, nil
	}
	return &report, nil
}
