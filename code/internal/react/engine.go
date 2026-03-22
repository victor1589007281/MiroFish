package react

import (
	"context"
	"encoding/json"
	"fmt"

	"swarm-predict/internal/model"
	"swarm-predict/internal/openclaw"
)

const maxReACTIterations = 10

// Engine implements the ReACT (Reasoning + Acting) loop for report generation.
type Engine struct {
	cc    openclaw.ChatCompleter
	tools []Tool
	model string
}

func NewEngine(cc openclaw.ChatCompleter, llmModel string, tools ...Tool) *Engine {
	return &Engine{cc: cc, tools: tools, model: llmModel}
}

// GenerateReport produces a structured prediction report using ReACT.
func (e *Engine) GenerateReport(ctx context.Context, projectDesc string, simSummary string, predictions []model.PredictionResult) (*model.Report, error) {
	// Phase 1: Plan the report outline
	outline, err := e.planOutline(ctx, projectDesc, simSummary, predictions)
	if err != nil {
		return nil, fmt.Errorf("plan outline: %w", err)
	}

	// Phase 2: Generate each section via ReACT loop
	var sections []model.ReportSection
	for i, title := range outline {
		content, err := e.generateSection(ctx, title, projectDesc, simSummary, predictions, sections)
		if err != nil {
			content = fmt.Sprintf("Error generating section: %v", err)
		}
		sections = append(sections, model.ReportSection{
			Title:   title,
			Content: content,
			Order:   i + 1,
		})
	}

	return &model.Report{
		Sections:    sections,
		Predictions: predictions,
	}, nil
}

func (e *Engine) planOutline(ctx context.Context, projectDesc, simSummary string, predictions []model.PredictionResult) ([]string, error) {
	predSummary := ""
	for _, p := range predictions {
		predSummary += fmt.Sprintf("- %s → %.0f%% (confidence: %.0f%%)\n", p.Question, p.Probability*100, p.Confidence*100)
	}

	resp, err := e.cc.Complete(ctx, openclaw.ChatRequest{
		Model: e.model,
		Messages: []openclaw.ChatMessage{
			{Role: "system", Content: outlineSystemPrompt},
			{Role: "user", Content: fmt.Sprintf("Project: %s\n\nSimulation summary:\n%s\n\nPrediction results:\n%s\n\nGenerate a report outline as a JSON array of section titles.", projectDesc, simSummary, predSummary)},
		},
	})
	if err != nil {
		return defaultOutline(), nil
	}

	var outline []string
	if err := extractJSONArray(resp, &outline); err != nil {
		return defaultOutline(), nil
	}
	return outline, nil
}

func (e *Engine) generateSection(ctx context.Context, sectionTitle, projectDesc, simSummary string, predictions []model.PredictionResult, prevSections []model.ReportSection) (string, error) {
	messages := []openclaw.ChatMessage{
		{Role: "system", Content: sectionSystemPrompt},
		{Role: "user", Content: fmt.Sprintf("Write section: \"%s\"\n\nProject: %s\nSimulation: %s", sectionTitle, projectDesc, simSummary)},
	}

	for iter := 0; iter < maxReACTIterations; iter++ {
		resp, err := e.cc.Complete(ctx, openclaw.ChatRequest{
			Model:    e.model,
			Messages: messages,
			Tools:    toolDefs(e.tools),
		})
		if err != nil {
			return "", fmt.Errorf("react iteration %d: %w", iter, err)
		}

		if len(resp.Choices) == 0 {
			return "", fmt.Errorf("empty response at iteration %d", iter)
		}

		choice := resp.Choices[0]

		// If LLM returns content (final answer), we're done
		if choice.FinishReason == "stop" && choice.Message.Content != "" {
			return choice.Message.Content, nil
		}

		// If LLM requests tool calls, execute them
		if len(choice.Message.ToolCalls) > 0 {
			messages = append(messages, choice.Message)

			for _, tc := range choice.Message.ToolCalls {
				tool := findTool(e.tools, tc.Function.Name)
				if tool == nil {
					messages = append(messages, openclaw.ChatMessage{
						Role:       "tool",
						Content:    fmt.Sprintf("Unknown tool: %s", tc.Function.Name),
						ToolCallID: tc.ID,
					})
					continue
				}
				result, err := tool.Execute(ctx, tc.Function.Arguments)
				if err != nil {
					result = fmt.Sprintf("Tool error: %v", err)
				}
				messages = append(messages, openclaw.ChatMessage{
					Role:       "tool",
					Content:    result,
					ToolCallID: tc.ID,
				})
			}
			continue
		}

		// Fallback: return whatever content we have
		if choice.Message.Content != "" {
			return choice.Message.Content, nil
		}
	}

	return "", fmt.Errorf("max ReACT iterations reached")
}

// Chat handles follow-up questions about the report.
func (e *Engine) Chat(ctx context.Context, report *model.Report, question string) (string, error) {
	reportText := ""
	for _, s := range report.Sections {
		reportText += fmt.Sprintf("## %s\n%s\n\n", s.Title, s.Content)
	}

	resp, err := e.cc.Complete(ctx, openclaw.ChatRequest{
		Model: e.model,
		Messages: []openclaw.ChatMessage{
			{Role: "system", Content: chatSystemPrompt},
			{Role: "user", Content: fmt.Sprintf("Report:\n%s\n\nQuestion: %s", reportText, question)},
		},
		Tools: toolDefs(e.tools),
	})
	if err != nil {
		return "", err
	}

	if len(resp.Choices) > 0 {
		return resp.Choices[0].Message.Content, nil
	}
	return "", fmt.Errorf("empty response")
}

func defaultOutline() []string {
	return []string{
		"Executive Summary",
		"Simulation Overview",
		"Key Findings",
		"Prediction Analysis",
		"Risk Assessment",
		"Conclusion & Recommendations",
	}
}

func extractJSONArray(resp *openclaw.ChatResponse, target *[]string) error {
	if resp == nil || len(resp.Choices) == 0 {
		return fmt.Errorf("empty response")
	}
	content := resp.Choices[0].Message.Content
	return parseJSONFromContent(content, target)
}

func parseJSONFromContent(content string, target interface{}) error {
	data := []byte(content)
	if err := json.Unmarshal(data, target); err == nil {
		return nil
	}
	s := content
	for _, prefix := range []string{"```json\n", "```\n", "```json", "```"} {
		if len(s) > len(prefix) && s[:len(prefix)] == prefix {
			s = s[len(prefix):]
			break
		}
	}
	if len(s) > 3 && s[len(s)-3:] == "```" {
		s = s[:len(s)-3]
	}
	return json.Unmarshal([]byte(s), target)
}

const outlineSystemPrompt = `You are a report planner. Given a project description, simulation results, and predictions, generate a report outline. Output a JSON array of section titles (5-8 sections). Example: ["Executive Summary", "Methodology", "Key Findings", "Predictions", "Risks", "Recommendations"]`

const sectionSystemPrompt = `You are a report writer for a prediction engine. Write a detailed section based on the title and available data. You can use tools to search for more information. Write in clear, analytical prose with data-backed claims.`

const chatSystemPrompt = `You are an AI assistant that helps users understand a prediction report. Answer questions about the report's findings, methodology, and predictions. Be specific and cite relevant data from the report.`
