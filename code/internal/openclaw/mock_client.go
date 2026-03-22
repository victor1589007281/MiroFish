package openclaw

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MockChatClient implements ChatCompleter for testing.
type MockChatClient struct {
	mu        sync.Mutex
	responses []ChatResponse
	calls     []ChatRequest
	handler   func(ChatRequest) (*ChatResponse, error)
}

func NewMockChatClient() *MockChatClient {
	return &MockChatClient{}
}

// SetHandler sets a custom handler function for dynamic responses.
func (m *MockChatClient) SetHandler(h func(ChatRequest) (*ChatResponse, error)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handler = h
}

// PushResponse enqueues a canned response (FIFO).
func (m *MockChatClient) PushResponse(resp ChatResponse) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responses = append(m.responses, resp)
}

// Calls returns all captured requests.
func (m *MockChatClient) Calls() []ChatRequest {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]ChatRequest{}, m.calls...)
}

func (m *MockChatClient) Complete(_ context.Context, req ChatRequest) (*ChatResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, req)

	if m.handler != nil {
		m.mu.Unlock()
		resp, err := m.handler(req)
		m.mu.Lock()
		return resp, err
	}

	if len(m.responses) == 0 {
		return SimpleTextResponse("mock response"), nil
	}
	resp := m.responses[0]
	m.responses = m.responses[1:]
	return &resp, nil
}

// MockSpawnClient implements Spawner for testing.
type MockSpawnClient struct {
	mu      sync.Mutex
	results map[string]*SubagentResult
	calls   []SpawnRequest
	handler func(SpawnRequest) (string, string)
}

func NewMockSpawnClient() *MockSpawnClient {
	return &MockSpawnClient{
		results: make(map[string]*SubagentResult),
	}
}

// SetHandler provides a function that receives a SpawnRequest and returns (runID, output).
func (m *MockSpawnClient) SetHandler(h func(SpawnRequest) (string, string)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handler = h
}

// SetResult pre-configures a result for a given runID.
func (m *MockSpawnClient) SetResult(runID string, output string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.results[runID] = &SubagentResult{
		RunID:  runID,
		Status: "completed",
		Output: output,
	}
}

func (m *MockSpawnClient) Calls() []SpawnRequest {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]SpawnRequest{}, m.calls...)
}

func (m *MockSpawnClient) Spawn(_ context.Context, req SpawnRequest) (*SpawnResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, req)

	runID := uuid.New().String()
	output := fmt.Sprintf(`{"decisions": [{"action": "post", "content": "mock output for %s"}]}`, req.Label)

	if m.handler != nil {
		m.mu.Unlock()
		runID, output = m.handler(req)
		m.mu.Lock()
	}

	m.results[runID] = &SubagentResult{
		RunID:  runID,
		Status: "completed",
		Output: output,
	}

	sessionKey := runID
	if len(sessionKey) > 8 {
		sessionKey = sessionKey[:8]
	}
	return &SpawnResponse{
		Status:          "accepted",
		RunID:           runID,
		ChildSessionKey: "mock-session-" + sessionKey,
	}, nil
}

func (m *MockSpawnClient) Poll(_ context.Context, runID string) (*SubagentResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if result, ok := m.results[runID]; ok {
		return result, nil
	}
	return &SubagentResult{RunID: runID, Status: "running"}, nil
}

func (m *MockSpawnClient) WaitForResult(ctx context.Context, runID string, _ time.Duration) (*SubagentResult, error) {
	return m.Poll(ctx, runID)
}

// --- Helpers ---

func SimpleTextResponse(text string) *ChatResponse {
	return &ChatResponse{
		ID: uuid.New().String(),
		Choices: []ChatChoice{
			{
				Index:        0,
				Message:      ChatMessage{Role: "assistant", Content: text},
				FinishReason: "stop",
			},
		},
	}
}

func ToolCallResponse(toolName, argsJSON string) *ChatResponse {
	return &ChatResponse{
		ID: uuid.New().String(),
		Choices: []ChatChoice{
			{
				Index: 0,
				Message: ChatMessage{
					Role: "assistant",
					ToolCalls: []ToolCall{
						{
							ID:   "call_" + uuid.New().String()[:8],
							Type: "function",
							Function: FunctionCall{
								Name:      toolName,
								Arguments: argsJSON,
							},
						},
					},
				},
				FinishReason: "tool_calls",
			},
		},
	}
}

func JSONTextResponse(v interface{}) *ChatResponse {
	data, _ := json.Marshal(v)
	return SimpleTextResponse(string(data))
}
