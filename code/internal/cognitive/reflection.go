package cognitive

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"swarm-predict/internal/model"
	"swarm-predict/internal/openclaw"
)

const (
	DefaultReflectionThreshold = 10.0
	reflectionHighImportance   = 0.9
)

// ReflectionEngine generates high-level insights from recent memories.
type ReflectionEngine struct {
	cc        openclaw.ChatCompleter
	model     string
	threshold float64
}

func NewReflectionEngine(cc openclaw.ChatCompleter, llmModel string) *ReflectionEngine {
	return &ReflectionEngine{
		cc:        cc,
		model:     llmModel,
		threshold: DefaultReflectionThreshold,
	}
}

func (r *ReflectionEngine) SetThreshold(t float64) { r.threshold = t }

// MaybeReflect checks if the agent has accumulated enough important memories
// since the last reflection. If so, it generates insights and writes them back.
func (r *ReflectionEngine) MaybeReflect(ctx context.Context, ms *MemoryStream, lastReflection time.Time) ([]model.MemoryEntry, error) {
	recent, err := ms.GetSince(ctx, lastReflection)
	if err != nil {
		return nil, fmt.Errorf("get recent memories: %w", err)
	}
	if SumImportance(recent) < r.threshold {
		return nil, nil
	}

	questions, err := r.generateQuestions(ctx, recent)
	if err != nil {
		return nil, fmt.Errorf("generate questions: %w", err)
	}

	var insights []model.MemoryEntry
	for _, q := range questions {
		relevant, err := ms.Retrieve(ctx, q, 10)
		if err != nil {
			continue
		}
		insight, err := r.generateInsight(ctx, q, relevant)
		if err != nil {
			continue
		}
		entry, err := ms.Add(ctx, insight, "reflection", reflectionHighImportance)
		if err != nil {
			continue
		}
		insights = append(insights, *entry)
	}
	return insights, nil
}

func (r *ReflectionEngine) generateQuestions(ctx context.Context, memories []model.MemoryEntry) ([]string, error) {
	memText := FormatMemories(memories)
	resp, err := r.cc.Complete(ctx, openclaw.ChatRequest{
		Model: r.model,
		Messages: []openclaw.ChatMessage{
			{Role: "system", Content: reflectionQuestionSystemPrompt},
			{Role: "user", Content: memText},
		},
	})
	if err != nil {
		return nil, err
	}
	content := ExtractContent(resp)
	return parseNumberedList(content), nil
}

func (r *ReflectionEngine) generateInsight(ctx context.Context, question string, memories []model.MemoryEntry) (string, error) {
	prompt := fmt.Sprintf("Question: %s\n\nRelevant memories:\n%s", question, FormatMemories(memories))
	resp, err := r.cc.Complete(ctx, openclaw.ChatRequest{
		Model: r.model,
		Messages: []openclaw.ChatMessage{
			{Role: "system", Content: reflectionInsightSystemPrompt},
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		return "", err
	}
	return ExtractContent(resp), nil
}

// FormatMemories renders entries for LLM context.
func FormatMemories(entries []model.MemoryEntry) string {
	var sb strings.Builder
	for i, e := range entries {
		fmt.Fprintf(&sb, "[%d] (%s, importance=%.1f) %s\n", i+1, e.Kind, e.Importance, e.Content)
	}
	return sb.String()
}

// ExtractContent pulls the text from the first choice.
func ExtractContent(resp *openclaw.ChatResponse) string {
	if resp == nil || len(resp.Choices) == 0 {
		return ""
	}
	return resp.Choices[0].Message.Content
}

// ExtractJSON pulls JSON from an LLM response into the target.
func ExtractJSON(resp *openclaw.ChatResponse, target interface{}) error {
	content := ExtractContent(resp)
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)
	return json.Unmarshal([]byte(content), target)
}

func parseNumberedList(text string) []string {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	var result []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		for i, ch := range line {
			if ch == '.' || ch == ')' || ch == '、' {
				line = strings.TrimSpace(line[i+1:])
				break
			}
			if i > 3 {
				break
			}
		}
		if line != "" {
			result = append(result, line)
		}
	}
	return result
}

const reflectionQuestionSystemPrompt = `You are analyzing an agent's recent experiences. Based on the memories provided, generate 3 high-level questions that would help the agent form deeper insights about patterns, relationships, or implications. Output as a numbered list.`

const reflectionInsightSystemPrompt = `You are synthesizing an insight for a simulated social agent. Given the question and relevant memories, produce a single concise insight (1-2 sentences) that captures a higher-level understanding. The insight should be something the agent learned or realized, not just a summary.`
