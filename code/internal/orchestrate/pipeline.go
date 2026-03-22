package orchestrate

import (
	"context"
	"fmt"
	"time"

	"swarm-predict/internal/cognitive"
	"swarm-predict/internal/model"
	"swarm-predict/internal/openclaw"
	"swarm-predict/internal/prediction"
	"swarm-predict/internal/react"
	"swarm-predict/internal/simulation"
)

// Config parameterizes the full pipeline.
type Config struct {
	Model       string
	FlashModel  string
	SimRounds   int
	AgentsCount int
	GroupSize   int
}

// PipelineResult is the full output of the six-step pipeline.
type PipelineResult struct {
	Agents         []model.AgentProfile            `json:"agents"`
	SimResult      *simulation.SimulationResult     `json:"simulation_result"`
	Emergence      simulation.EmergenceSignal       `json:"emergence"`
	RoundEmergence []simulation.EmergenceSignal     `json:"round_emergence,omitempty"`
	Predictions    []model.PredictionResult         `json:"predictions"`
	Report         *model.Report                    `json:"report"`
}

// Pipeline coordinates the full six-step MiroFish V2 workflow.
type Pipeline struct {
	cc      openclaw.ChatCompleter
	spawner openclaw.Spawner
	persona *simulation.PersonaGenerator
	config  Config
}

func NewPipeline(cc openclaw.ChatCompleter, spawner openclaw.Spawner, config Config) *Pipeline {
	return &Pipeline{
		cc:      cc,
		spawner: spawner,
		persona: simulation.NewPersonaGenerator(cc, config.Model),
		config:  config,
	}
}

// Run executes all six steps of the prediction pipeline.
func (p *Pipeline) Run(ctx context.Context, topic string, questions []string) (*PipelineResult, error) {
	result := &PipelineResult{}

	// Step 2: Generate agent personas
	agents, err := p.persona.Generate(ctx, topic, p.config.AgentsCount)
	if err != nil {
		return nil, fmt.Errorf("step2 persona generation: %w", err)
	}
	result.Agents = agents

	// Step 3: Run simulation with social graph for inter-agent perception tracking
	socialStore := cognitive.NewInMemorySocialStore()
	socialGraph := cognitive.NewSocialGraph(socialStore)

	simConfig := model.SimulationConfig{
		ID:             fmt.Sprintf("sim-%d", time.Now().Unix()),
		Rounds:         p.config.SimRounds,
		AgentsPerGroup: p.config.GroupSize,
		Model:          p.config.Model,
		FlashModel:     p.config.FlashModel,
	}
	simEngine := simulation.NewEngine(p.spawner, p.cc, simConfig,
		simulation.WithSocialGraph(socialGraph),
	)
	simResult, err := simEngine.Run(ctx, agents)
	if err != nil {
		return nil, fmt.Errorf("step3 simulation: %w", err)
	}
	result.SimResult = simResult
	result.RoundEmergence = simResult.RoundEmergence

	// Step 3.5: Final aggregate emergence detection across all rounds
	result.Emergence = simulation.DetectEmergence(simResult.ActionLogs)

	// Step 4: Prediction refinement — include emergence signals as context
	emergenceSummary := formatEmergenceForPrediction(result.Emergence)
	simSummary := react.SummarizeSimulation(simResult.TotalRounds, simResult.TotalPosts, nil, nil)
	enrichedSummary := simSummary + "\n\n" + emergenceSummary

	delphi := prediction.NewDelphiEngine(p.spawner, p.cc, p.config.Model)
	predictions, err := delphi.Refine(ctx, enrichedSummary, questions)
	if err != nil {
		return nil, fmt.Errorf("step4 prediction: %w", err)
	}
	result.Predictions = predictions

	// Step 5: Report generation with full tool suite
	reactTools := []react.Tool{
		react.NewSimDataTool(simResult.ActionLogs),
		react.NewEmergenceTool(simResult.ActionLogs),
		react.NewPredictionTool(predictions),
	}
	reportEngine := react.NewEngine(p.cc, p.config.Model, reactTools...)
	report, err := reportEngine.GenerateReport(ctx, topic, enrichedSummary, predictions)
	if err != nil {
		return nil, fmt.Errorf("step5 report: %w", err)
	}
	result.Report = report

	return result, nil
}

func formatEmergenceForPrediction(signal simulation.EmergenceSignal) string {
	summary := fmt.Sprintf("Emergence Analysis:\n- Polarization Index: %.2f\n- Consensus Level: %.2f",
		signal.PolarizationIndex, signal.ConsensusLevel)
	if len(signal.TopTopics) > 0 {
		summary += fmt.Sprintf("\n- Top Topics: %v", signal.TopTopics)
	}
	if len(signal.InfluencerShifts) > 0 {
		summary += fmt.Sprintf("\n- Influencer Shifts: %v", signal.InfluencerShifts)
	}
	return summary
}
