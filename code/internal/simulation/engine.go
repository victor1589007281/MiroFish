package simulation

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"swarm-predict/internal/cognitive"
	"swarm-predict/internal/model"
	"swarm-predict/internal/openclaw"

	"golang.org/x/sync/errgroup"
)

// Engine orchestrates the multi-agent simulation.
type Engine struct {
	spawner  openclaw.Spawner
	cc       openclaw.ChatCompleter
	gm       *GameMaster
	platform *Platform
	feed     *FeedAlgorithm
	social   *cognitive.SocialGraph
	memories map[string]*cognitive.MemoryStream
	logger   *Logger
	config   model.SimulationConfig

	// Per-round emergence signals
	RoundEmergence []EmergenceSignal
}

// EngineOption allows optional injection of dependencies.
type EngineOption func(*Engine)

func WithSocialGraph(sg *cognitive.SocialGraph) EngineOption {
	return func(e *Engine) { e.social = sg }
}

func WithMemoryStreams(ms map[string]*cognitive.MemoryStream) EngineOption {
	return func(e *Engine) { e.memories = ms }
}

func NewEngine(spawner openclaw.Spawner, cc openclaw.ChatCompleter, config model.SimulationConfig, opts ...EngineOption) *Engine {
	plat := NewPlatform()
	e := &Engine{
		spawner:  spawner,
		cc:       cc,
		gm:       NewGameMaster(cc, config.Model),
		platform: plat,
		feed:     NewFeedAlgorithm(plat),
		memories: make(map[string]*cognitive.MemoryStream),
		logger:   NewInMemoryLogger(),
		config:   config,
	}
	for _, o := range opts {
		o(e)
	}
	return e
}

func (e *Engine) SetLogger(l *Logger)    { e.logger = l }
func (e *Engine) Platform() *Platform    { return e.platform }
func (e *Engine) Logger() *Logger        { return e.logger }

// Run executes the full simulation.
func (e *Engine) Run(ctx context.Context, agents []model.AgentProfile) (*SimulationResult, error) {
	groups := GroupAgents(agents, e.config.AgentsPerGroup)
	profileMap := make(map[string]model.AgentProfile)
	for _, a := range agents {
		profileMap[a.ID] = a
	}

	// Initialize memory streams for agents if not already set
	memStore := cognitive.NewInMemoryStore()
	for _, a := range agents {
		if _, ok := e.memories[a.ID]; !ok {
			e.memories[a.ID] = cognitive.NewMemoryStream(a.ID, memStore, nil)
		}
	}

	for round := 1; round <= e.config.Rounds; round++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		for _, evt := range e.config.Events {
			if evt.Round == round {
				e.platform.CreatePost("system", evt.Content)
			}
		}

		if err := e.runRound(ctx, round, groups, profileMap); err != nil {
			return nil, fmt.Errorf("round %d: %w", round, err)
		}
	}

	return &SimulationResult{
		TotalRounds:    e.config.Rounds,
		TotalPosts:     e.platform.PostCount(),
		ActionLogs:     e.logger.GetLogs(),
		RoundEmergence: e.RoundEmergence,
	}, nil
}

// SimulationResult holds the outcome of a completed simulation.
type SimulationResult struct {
	TotalRounds    int               `json:"total_rounds"`
	TotalPosts     int               `json:"total_posts"`
	ActionLogs     []ActionLog       `json:"action_logs"`
	RoundEmergence []EmergenceSignal `json:"round_emergence,omitempty"`
}

func (e *Engine) runRound(ctx context.Context, round int, groups [][]model.AgentProfile, profiles map[string]model.AgentProfile) error {
	// --- Phase 1: Spawn cognition for each group (concurrent) ---
	var mu sync.Mutex
	var allProposals []model.ActionProposal

	g, gctx := errgroup.WithContext(ctx)
	for gi, group := range groups {
		gi, group := gi, group
		g.Go(func() error {
			proposals, err := e.spawnGroupCognition(gctx, round, gi, group)
			if err != nil {
				return err
			}
			mu.Lock()
			allProposals = append(allProposals, proposals...)
			mu.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return fmt.Errorf("group cognition: %w", err)
	}

	// --- Phase 2: GM arbitrates all proposals ---
	agentMemories := make(map[string][]model.MemoryEntry)
	for agentID, ms := range e.memories {
		mems, _ := ms.Retrieve(ctx, "recent actions", 5)
		agentMemories[agentID] = mems
	}
	decisions, err := e.gm.ArbitrateBatch(ctx, allProposals, profiles, agentMemories)
	if err != nil {
		return fmt.Errorf("gm arbitration: %w", err)
	}

	// --- Phase 3: Execute + Update Social Graph + Write Memory ---
	for i := range decisions {
		decisions[i].Round = round
		e.platform.Execute(decisions[i])
		e.logger.Log(round, decisions[i])

		// Write action result back to agent's memory stream
		if ms, ok := e.memories[decisions[i].AgentID]; ok {
			desc := fmt.Sprintf("Round %d: I %s", round, describeAction(decisions[i]))
			ms.Add(ctx, desc, "action", 0.6)
		}

		// Update social graph after interaction
		if e.social != nil && decisions[i].TargetID != "" {
			delta := inferSocialDelta(decisions[i].ActionType)
			e.social.UpdateAfterInteraction(ctx, decisions[i].AgentID, decisions[i].TargetID, delta)
		}
	}

	// --- Phase 4: Per-round emergence detection ---
	roundLogs := e.logger.GetLogsByRound(round)
	signal := DetectEmergence(roundLogs)
	e.RoundEmergence = append(e.RoundEmergence, signal)

	return nil
}

func (e *Engine) spawnGroupCognition(ctx context.Context, round, groupIdx int, agents []model.AgentProfile) ([]model.ActionProposal, error) {
	// Build enriched prompt with personalized feed and memory context per agent
	prompt := e.buildEnrichedCognitionPrompt(ctx, agents, round)

	resp, err := e.spawner.Spawn(ctx, openclaw.SpawnRequest{
		Task:              prompt,
		Model:             e.config.FlashModel,
		Label:             fmt.Sprintf("sim-r%d-g%d", round, groupIdx),
		RunTimeoutSeconds: 120,
		Deliver:           true,
	})
	if err != nil {
		return nil, fmt.Errorf("spawn group %d: %w", groupIdx, err)
	}

	result, err := e.spawner.WaitForResult(ctx, resp.RunID, 2*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("wait group %d: %w", groupIdx, err)
	}

	return parseGroupDecisions(result.Output, agents), nil
}

// buildEnrichedCognitionPrompt includes memory, reflections, social context per agent.
func (e *Engine) buildEnrichedCognitionPrompt(ctx context.Context, agents []model.AgentProfile, round int) string {
	agentDescs := ""
	for _, a := range agents {
		desc := fmt.Sprintf("- %s (ID: %s, Role: %s): %s\n", a.Name, a.ID, a.Role, a.Personality)

		// Attach memory context
		if ms, ok := e.memories[a.ID]; ok {
			mems, err := ms.Retrieve(ctx, a.Role+" "+a.Personality, 5)
			if err == nil && len(mems) > 0 {
				desc += "  Recent memories:\n"
				for _, m := range mems {
					desc += fmt.Sprintf("    - [%s] %s\n", m.Kind, m.Content)
				}
			}
		}

		// Attach social context
		if e.social != nil {
			rels, err := e.social.GetRelationships(ctx, a.ID)
			if err == nil && len(rels) > 0 {
				desc += "  Social relationships:\n"
				for _, rel := range rels {
					if len(desc) > 2000 {
						break
					}
					desc += fmt.Sprintf("    - %s (like=%.1f, rep=%.1f)\n", rel.TargetID, rel.Likability, rel.Reputation)
				}
			}
		}

		agentDescs += desc
	}

	// Use personalized feed for the group (based on first agent's profile)
	var feedStr string
	feed := e.feed.PersonalizedFeed(agents[0], nil, 10)
	if len(feed) == 0 {
		feed = e.platform.GetFeed(10)
	}
	for _, p := range feed {
		feedStr += fmt.Sprintf("- @%s: %s (likes: %d)\n", p.AuthorID, p.Content, p.Likes)
	}

	return fmt.Sprintf(`Round %d of social simulation.

Agents in this group (with their memory and social context):
%s

Current information feed:
%s

For EACH agent, consider their personality, memories, social relationships, and the current feed.
Decide their action this round following their persona consistently.
Output JSON array: [{"agent_id": "...", "action_type": "post|reply|like|repost|follow", "content": "...", "target_id": "..."}]`, round, agentDescs, feedStr)
}

func parseGroupDecisions(output string, agents []model.AgentProfile) []model.ActionProposal {
	var raw []struct {
		AgentID    string `json:"agent_id"`
		ActionType string `json:"action_type"`
		Content    string `json:"content"`
		TargetID   string `json:"target_id"`
	}

	if err := json.Unmarshal([]byte(output), &raw); err != nil {
		var proposals []model.ActionProposal
		for _, a := range agents {
			proposals = append(proposals, model.ActionProposal{
				AgentID:    a.ID,
				ActionType: "like",
				Intent:     "parse error fallback",
			})
		}
		return proposals
	}

	proposals := make([]model.ActionProposal, 0, len(raw))
	for _, r := range raw {
		proposals = append(proposals, model.ActionProposal{
			AgentID:    r.AgentID,
			Intent:     fmt.Sprintf("%s wants to %s", r.AgentID, r.ActionType),
			ActionType: r.ActionType,
			Content:    r.Content,
			TargetID:   r.TargetID,
		})
	}
	return proposals
}

// describeAction generates a natural-language description for memory write-back.
func describeAction(d model.Decision) string {
	switch d.ActionType {
	case "post":
		return fmt.Sprintf("posted: \"%s\"", truncate(d.Content, 80))
	case "reply":
		return fmt.Sprintf("replied to %s: \"%s\"", d.TargetID, truncate(d.Content, 60))
	case "like":
		return fmt.Sprintf("liked post %s", d.TargetID)
	case "repost":
		return fmt.Sprintf("reposted %s", d.TargetID)
	case "follow":
		return fmt.Sprintf("followed %s", d.TargetID)
	default:
		return fmt.Sprintf("did %s", d.ActionType)
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// inferSocialDelta determines social perception change from an action type.
func inferSocialDelta(actionType string) cognitive.InteractionDelta {
	switch actionType {
	case "like":
		return cognitive.InteractionDelta{LikabilityDelta: 0.05, InfluenceDelta: 0.01}
	case "reply":
		return cognitive.InteractionDelta{LikabilityDelta: 0.1, ReputationDelta: 0.05}
	case "repost":
		return cognitive.InteractionDelta{LikabilityDelta: 0.08, InfluenceDelta: 0.05}
	case "follow":
		return cognitive.InteractionDelta{LikabilityDelta: 0.15, InfluenceDelta: 0.03}
	default:
		return cognitive.InteractionDelta{}
	}
}
