package openclaw

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ChatMessage represents a message in the OpenAI-compatible chat format.
type ChatMessage struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

// ToolCall from an LLM response.
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ToolDef defines a tool the LLM can call.
type ToolDef struct {
	Type     string         `json:"type"`
	Function ToolFuncDef    `json:"function"`
}

type ToolFuncDef struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
}

// ChatRequest is the request body for chat completions.
type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Tools    []ToolDef     `json:"tools,omitempty"`
}

// ChatResponse is the response from chat completions.
type ChatResponse struct {
	ID      string         `json:"id"`
	Choices []ChatChoice   `json:"choices"`
	Usage   *UsageInfo     `json:"usage,omitempty"`
}

type ChatChoice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

type UsageInfo struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// SpawnRequest is the request to spawn a subagent via tools/invoke.
type SpawnRequest struct {
	Task              string `json:"task"`
	Model             string `json:"model,omitempty"`
	Label             string `json:"label,omitempty"`
	RunTimeoutSeconds int    `json:"runTimeoutSeconds,omitempty"`
	Deliver           bool   `json:"deliver,omitempty"`
}

// SpawnResponse from sessions_spawn.
type SpawnResponse struct {
	Status          string `json:"status"`
	RunID           string `json:"runId"`
	ChildSessionKey string `json:"childSessionKey"`
}

// SubagentResult holds the completed subagent's output.
type SubagentResult struct {
	RunID   string `json:"run_id"`
	Status  string `json:"status"` // completed | failed | timeout
	Output  string `json:"output"`
}

// ChatCompleter is the interface for OpenClaw chat completions (Main Lane).
type ChatCompleter interface {
	Complete(ctx context.Context, req ChatRequest) (*ChatResponse, error)
}

// Spawner is the interface for spawning subagents via toolsinvoke (Subagent Lane).
type Spawner interface {
	Spawn(ctx context.Context, req SpawnRequest) (*SpawnResponse, error)
	Poll(ctx context.Context, runID string) (*SubagentResult, error)
	WaitForResult(ctx context.Context, runID string, timeout time.Duration) (*SubagentResult, error)
}

// --- Real implementations ---

// HTTPChatClient calls the OpenClaw Gateway /v1/chat/completions endpoint.
type HTTPChatClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewHTTPChatClient(baseURL, apiKey string) *HTTPChatClient {
	return &HTTPChatClient{
		baseURL:    baseURL,
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 120 * time.Second},
	}
}

func (c *HTTPChatClient) Complete(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("chat completions returned %d: %s", resp.StatusCode, string(respBody))
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &chatResp, nil
}

// HTTPSpawnClient calls the OpenClaw Gateway /tools/invoke endpoint.
type HTTPSpawnClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewHTTPSpawnClient(baseURL, apiKey string) *HTTPSpawnClient {
	return &HTTPSpawnClient{
		baseURL:    baseURL,
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *HTTPSpawnClient) Spawn(ctx context.Context, req SpawnRequest) (*SpawnResponse, error) {
	invokeReq := map[string]interface{}{
		"tool": "sessions_spawn",
		"args": req,
	}
	body, err := json.Marshal(invokeReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/tools/invoke", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("tools/invoke returned %d: %s", resp.StatusCode, string(respBody))
	}

	var spawnResp SpawnResponse
	if err := json.NewDecoder(resp.Body).Decode(&spawnResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &spawnResp, nil
}

func (c *HTTPSpawnClient) Poll(ctx context.Context, runID string) (*SubagentResult, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/tools/invoke?tool=sessions_list&runId=%s", c.baseURL, runID), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	var result SubagentResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	result.RunID = runID
	return &result, nil
}

func (c *HTTPSpawnClient) WaitForResult(ctx context.Context, runID string, timeout time.Duration) (*SubagentResult, error) {
	deadline := time.Now().Add(timeout)
	interval := 2 * time.Second

	for time.Now().Before(deadline) {
		result, err := c.Poll(ctx, runID)
		if err != nil {
			return nil, err
		}
		if result.Status == "completed" || result.Status == "failed" {
			return result, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(interval):
			if interval < 10*time.Second {
				interval = interval * 3 / 2
			}
		}
	}

	return nil, fmt.Errorf("timeout waiting for subagent %s after %v", runID, timeout)
}
