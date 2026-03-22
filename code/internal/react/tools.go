package react

import (
	"context"
	"encoding/json"
	"fmt"

	"swarm-predict/internal/model"
	"swarm-predict/internal/openclaw"
	"swarm-predict/internal/simulation"
)

// Tool is a callable tool the ReACT engine can use.
type Tool struct {
	Definition openclaw.ToolDef
	Execute    func(ctx context.Context, args string) (string, error)
}

// NewGraphSearchTool creates a tool for searching the knowledge graph.
type GraphSearcher interface {
	Search(ctx context.Context, query string, limit int) ([]map[string]interface{}, error)
}

func NewGraphSearchTool(searcher GraphSearcher) Tool {
	return Tool{
		Definition: openclaw.ToolDef{
			Type: "function",
			Function: openclaw.ToolFuncDef{
				Name:        "graph_search",
				Description: "Search the knowledge graph for entities, relationships, and facts relevant to the query.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"query": map[string]interface{}{"type": "string", "description": "semantic search query"},
						"limit": map[string]interface{}{"type": "integer", "description": "max results (default 10)"},
					},
					"required": []string{"query"},
				},
			},
		},
		Execute: func(ctx context.Context, args string) (string, error) {
			var params struct {
				Query string `json:"query"`
				Limit int    `json:"limit"`
			}
			json.Unmarshal([]byte(args), &params)
			if params.Limit <= 0 {
				params.Limit = 10
			}
			results, err := searcher.Search(ctx, params.Query, params.Limit)
			if err != nil {
				return "", err
			}
			data, _ := json.Marshal(results)
			return string(data), nil
		},
	}
}

// NewSimDataTool creates a tool for querying simulation results.
func NewSimDataTool(logs []simulation.ActionLog) Tool {
	return Tool{
		Definition: openclaw.ToolDef{
			Type: "function",
			Function: openclaw.ToolFuncDef{
				Name:        "simulation_data",
				Description: "Query simulation action logs. Get posts, behaviors, and trends from the simulation.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"query":     map[string]interface{}{"type": "string", "description": "what to search for in simulation data"},
						"round_min": map[string]interface{}{"type": "integer", "description": "minimum round number"},
						"round_max": map[string]interface{}{"type": "integer", "description": "maximum round number"},
					},
					"required": []string{"query"},
				},
			},
		},
		Execute: func(_ context.Context, args string) (string, error) {
			var params struct {
				Query    string `json:"query"`
				RoundMin int    `json:"round_min"`
				RoundMax int    `json:"round_max"`
			}
			json.Unmarshal([]byte(args), &params)

			var filtered []simulation.ActionLog
			for _, log := range logs {
				if params.RoundMin > 0 && log.Round < params.RoundMin {
					continue
				}
				if params.RoundMax > 0 && log.Round > params.RoundMax {
					continue
				}
				filtered = append(filtered, log)
			}

			if len(filtered) > 50 {
				filtered = filtered[:50]
			}
			data, _ := json.Marshal(filtered)
			return string(data), nil
		},
	}
}

// NewEmergenceTool creates a tool for analyzing emergent patterns.
func NewEmergenceTool(logs []simulation.ActionLog) Tool {
	return Tool{
		Definition: openclaw.ToolDef{
			Type: "function",
			Function: openclaw.ToolFuncDef{
				Name:        "emergence_analysis",
				Description: "Analyze emergent patterns in the simulation: polarization, trending topics, consensus levels.",
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
		},
		Execute: func(_ context.Context, _ string) (string, error) {
			signal := simulation.DetectEmergence(logs)
			data, _ := json.Marshal(signal)
			return string(data), nil
		},
	}
}

// NewPredictionTool wraps prediction results as a queryable tool.
func NewPredictionTool(predictions []model.PredictionResult) Tool {
	return Tool{
		Definition: openclaw.ToolDef{
			Type: "function",
			Function: openclaw.ToolFuncDef{
				Name:        "prediction_results",
				Description: "Get the refined prediction results from the Delphi process.",
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
		},
		Execute: func(_ context.Context, _ string) (string, error) {
			data, _ := json.Marshal(predictions)
			return string(data), nil
		},
	}
}

func toolDefs(tools []Tool) []openclaw.ToolDef {
	defs := make([]openclaw.ToolDef, len(tools))
	for i, t := range tools {
		defs[i] = t.Definition
	}
	return defs
}

func findTool(tools []Tool, name string) *Tool {
	for _, t := range tools {
		if t.Definition.Function.Name == name {
			return &t
		}
	}
	return nil
}

func formatToolResult(name, result string) string {
	return fmt.Sprintf("[Tool: %s]\n%s", name, result)
}
