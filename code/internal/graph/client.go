package graph

import (
	"context"
)

// Client is the interface for knowledge graph operations.
type Client interface {
	AddEntity(ctx context.Context, entityType, name string, properties map[string]interface{}) error
	AddRelation(ctx context.Context, source, target, relType string, properties map[string]interface{}) error
	Search(ctx context.Context, query string, limit int) ([]map[string]interface{}, error)
	Close() error
}

// MockClient is a test-friendly in-memory graph.
type MockClient struct {
	entities  []map[string]interface{}
	relations []map[string]interface{}
}

func NewMockClient() *MockClient {
	return &MockClient{}
}

func (m *MockClient) AddEntity(_ context.Context, entityType, name string, properties map[string]interface{}) error {
	entry := map[string]interface{}{
		"type": entityType,
		"name": name,
	}
	for k, v := range properties {
		entry[k] = v
	}
	m.entities = append(m.entities, entry)
	return nil
}

func (m *MockClient) AddRelation(_ context.Context, source, target, relType string, properties map[string]interface{}) error {
	entry := map[string]interface{}{
		"source":   source,
		"target":   target,
		"rel_type": relType,
	}
	for k, v := range properties {
		entry[k] = v
	}
	m.relations = append(m.relations, entry)
	return nil
}

func (m *MockClient) Search(_ context.Context, query string, limit int) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	for _, e := range m.entities {
		if len(results) >= limit {
			break
		}
		results = append(results, e)
	}
	return results, nil
}

func (m *MockClient) Close() error { return nil }

func (m *MockClient) Entities() []map[string]interface{} { return m.entities }
func (m *MockClient) Relations() []map[string]interface{} { return m.relations }
