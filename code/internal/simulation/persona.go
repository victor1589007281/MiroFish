package simulation

import (
	"context"
	"fmt"

	"swarm-predict/internal/cognitive"
	"swarm-predict/internal/model"
	"swarm-predict/internal/openclaw"
)

// PersonaGenerator creates agent profiles from a topic description using LLM.
type PersonaGenerator struct {
	cc    openclaw.ChatCompleter
	model string
}

func NewPersonaGenerator(cc openclaw.ChatCompleter, llmModel string) *PersonaGenerator {
	return &PersonaGenerator{cc: cc, model: llmModel}
}

// Generate creates diverse agent profiles for a given prediction topic.
func (pg *PersonaGenerator) Generate(ctx context.Context, topic string, count int) ([]model.AgentProfile, error) {
	prompt := fmt.Sprintf(`Generate %d diverse social media user personas for simulating a discussion about: "%s"

Requirements:
- Each persona should have a unique perspective and role
- Include a mix of stakeholders: government officials, media, industry experts, general public, activists, academics
- Each persona should have distinct personality traits that influence their social media behavior
- Provide backgrounds that justify their positions

Output JSON array:
[{"id": "unique-id", "name": "...", "role": "...", "personality": "...", "background": "...", "goals": ["goal1", "goal2"]}]`, count, topic)

	resp, err := pg.cc.Complete(ctx, openclaw.ChatRequest{
		Model: pg.model,
		Messages: []openclaw.ChatMessage{
			{Role: "system", Content: personaSystemPrompt},
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("generate personas: %w", err)
	}

	var profiles []model.AgentProfile
	if err := cognitive.ExtractJSON(resp, &profiles); err != nil {
		return nil, fmt.Errorf("parse personas: %w", err)
	}

	// Ensure unique IDs
	for i := range profiles {
		if profiles[i].ID == "" {
			profiles[i].ID = fmt.Sprintf("agent-%d", i+1)
		}
	}

	return profiles, nil
}

// GenerateViaSpawn uses subagent spawn to generate personas in parallel groups.
func (pg *PersonaGenerator) GenerateViaSpawn(ctx context.Context, spawner openclaw.Spawner, topic string, count int) ([]model.AgentProfile, error) {
	groupSize := 6
	numGroups := (count + groupSize - 1) / groupSize

	var allProfiles []model.AgentProfile
	for g := 0; g < numGroups; g++ {
		n := groupSize
		if g == numGroups-1 && count%groupSize != 0 {
			n = count % groupSize
		}
		prompt := fmt.Sprintf(`Generate %d diverse social media user personas for group %d of a simulation about: "%s"
Output JSON array: [{"id": "g%d-N", "name": "...", "role": "...", "personality": "...", "background": "...", "goals": [...]}]`,
			n, g+1, topic, g+1)

		resp, err := spawner.Spawn(ctx, openclaw.SpawnRequest{
			Task:              prompt,
			Model:             pg.model,
			Label:             fmt.Sprintf("persona-g%d", g+1),
			RunTimeoutSeconds: 60,
			Deliver:           true,
		})
		if err != nil {
			continue
		}

		result, err := spawner.WaitForResult(ctx, resp.RunID, 60*1e9)
		if err != nil {
			continue
		}

		var profiles []model.AgentProfile
		if err := cognitive.ExtractJSON(openclaw.SimpleTextResponse(result.Output), &profiles); err == nil {
			allProfiles = append(allProfiles, profiles...)
		}
	}

	if len(allProfiles) == 0 {
		return nil, fmt.Errorf("failed to generate any personas")
	}
	return allProfiles, nil
}

const personaSystemPrompt = `You are an expert at creating diverse, realistic social media personas for simulation. 
Each persona should feel authentic with distinct viewpoints, communication styles, and motivations.
Ensure diversity across: age, profession, political leaning, expertise level, personality type.
Always output valid JSON.`
