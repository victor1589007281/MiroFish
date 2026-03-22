package simulation

import (
	"fmt"
	"os"

	"swarm-predict/internal/model"

	"gopkg.in/yaml.v3"
)

// ExpertsConfig maps to configs/experts.yaml.
type ExpertsConfig struct {
	Experts []model.ExpertPerspective `yaml:"experts"`
}

// LoadExperts reads expert configurations from a YAML file.
func LoadExperts(path string) ([]model.ExpertPerspective, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read experts config: %w", err)
	}

	var config ExpertsConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parse experts config: %w", err)
	}

	return config.Experts, nil
}

// LoadExpertsOrDefault tries to load from file, falls back to defaults.
func LoadExpertsOrDefault(path string) []model.ExpertPerspective {
	experts, err := LoadExperts(path)
	if err != nil || len(experts) == 0 {
		return defaultExpertsFallback()
	}
	return experts
}

func defaultExpertsFallback() []model.ExpertPerspective {
	return []model.ExpertPerspective{
		{Name: "optimist", SystemPrompt: "You are an optimistic analyst focusing on positive signals."},
		{Name: "pessimist", SystemPrompt: "You are a risk analyst focusing on downside scenarios."},
		{Name: "quant", SystemPrompt: "You are a quantitative analyst focusing on data-driven evidence."},
		{Name: "domain_expert", SystemPrompt: "You are a domain expert drawing on historical precedents."},
		{Name: "strategist", SystemPrompt: "You are a game theory expert analyzing stakeholder incentives."},
	}
}
