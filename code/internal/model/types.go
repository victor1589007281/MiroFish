package model

import "time"

// AgentProfile holds the persona and configuration of a simulation agent.
type AgentProfile struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Role        string            `json:"role"`
	Personality string            `json:"personality"`
	Background  string            `json:"background"`
	Goals       []string          `json:"goals"`
	Traits      map[string]string `json:"traits"`
}

// MemoryEntry is a single record in an agent's memory stream.
type MemoryEntry struct {
	ID         string    `json:"id"`
	AgentID    string    `json:"agent_id"`
	Content    string    `json:"content"`
	Timestamp  time.Time `json:"timestamp"`
	Importance float64   `json:"importance"`
	Kind       string    `json:"kind"` // observation | action | reflection | plan
	Embedding  []float32 `json:"embedding,omitempty"`
}

// SocialRelation represents one agent's perception of another.
type SocialRelation struct {
	ObserverID  string    `json:"observer_id"`
	TargetID    string    `json:"target_id"`
	Reputation  float64   `json:"reputation"`  // [-1, 1]
	Influence   float64   `json:"influence"`   // [0, 1]
	Likability  float64   `json:"likability"`  // [-1, 1]
	LastUpdated time.Time `json:"last_updated"`
}

// ActionProposal is what an agent wants to do, before GM arbitration.
type ActionProposal struct {
	AgentID    string `json:"agent_id"`
	Intent     string `json:"intent"`
	ActionType string `json:"action_type"` // post | reply | like | follow | repost
	Content    string `json:"content,omitempty"`
	TargetID   string `json:"target_id,omitempty"`
}

// GMVerdict is the Game Master's ruling on an ActionProposal.
type GMVerdict struct {
	Approved       bool            `json:"approved"`
	Reason         string          `json:"reason"`
	ModifiedAction *ActionProposal `json:"modified_action,omitempty"`
}

// Decision is the resolved action after GM arbitration, ready to execute.
type Decision struct {
	AgentID    string `json:"agent_id"`
	ActionType string `json:"action_type"`
	Content    string `json:"content"`
	TargetID   string `json:"target_id,omitempty"`
	Round      int    `json:"round"`
}

// Post represents a social media post in the simulated platform.
type Post struct {
	ID        string    `json:"id"`
	AuthorID  string    `json:"author_id"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Likes     int       `json:"likes"`
	Reposts   int       `json:"reposts"`
	Replies   []Reply   `json:"replies,omitempty"`
}

// Reply is a comment on a post.
type Reply struct {
	ID        string    `json:"id"`
	AuthorID  string    `json:"author_id"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// SimulationConfig holds runtime parameters for a simulation.
type SimulationConfig struct {
	ID          string         `json:"id"`
	ProjectID   string         `json:"project_id"`
	Rounds      int            `json:"rounds"`
	AgentsPerGroup int         `json:"agents_per_group"`
	Events      []EventConfig  `json:"events"`
	Model       string         `json:"model"`
	FlashModel  string         `json:"flash_model"`
}

// EventConfig defines an event to inject at a specific round.
type EventConfig struct {
	Round   int    `json:"round"`
	Content string `json:"content"`
}

// PredictionResult is the output of the Delphi prediction refinement.
type PredictionResult struct {
	Question      string             `json:"question"`
	Probability   float64            `json:"probability"`
	Confidence    float64            `json:"confidence"`
	KeyArguments  []string           `json:"key_arguments"`
	Disagreements []string           `json:"disagreements"`
	ExpertViews   map[string]float64 `json:"expert_views"`
	Rounds        int                `json:"rounds"`
}

// ExpertPerspective defines a Delphi expert agent's persona.
type ExpertPerspective struct {
	Name         string `json:"name" yaml:"name"`
	SystemPrompt string `json:"system_prompt" yaml:"system_prompt"`
}

// SimulationState tracks the overall state of a running simulation.
type SimulationState struct {
	ID           string    `json:"id"`
	Status       string    `json:"status"` // pending | running | completed | failed
	CurrentRound int       `json:"current_round"`
	TotalRounds  int       `json:"total_rounds"`
	StartedAt    time.Time `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	Error        string    `json:"error,omitempty"`
}

// ReportSection is one section of the final prediction report.
type ReportSection struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Order   int    `json:"order"`
}

// Report is the complete prediction report.
type Report struct {
	ID           string          `json:"id"`
	ProjectID    string          `json:"project_id"`
	SimulationID string          `json:"simulation_id"`
	Sections     []ReportSection `json:"sections"`
	Predictions  []PredictionResult `json:"predictions"`
	CreatedAt    time.Time       `json:"created_at"`
}

// OntologyResult from LLM analysis of seed documents.
type OntologyResult struct {
	EntityTypes   []EntityType   `json:"entity_types"`
	RelationTypes []RelationType `json:"relation_types"`
}

type EntityType struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type RelationType struct {
	Name        string `json:"name"`
	Source      string `json:"source"`
	Target      string `json:"target"`
	Description string `json:"description"`
}
