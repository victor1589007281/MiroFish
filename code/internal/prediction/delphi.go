package prediction

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"swarm-predict/internal/model"
	"swarm-predict/internal/openclaw"
)

const (
	DefaultMaxRounds          = 4
	DefaultConvergenceThreshold = 0.05
)

// DelphiEngine runs multi-round expert deliberation for prediction refinement.
type DelphiEngine struct {
	spawner              openclaw.Spawner
	cc                   openclaw.ChatCompleter
	mediator             *Mediator
	calibrator           *Calibrator
	experts              []model.ExpertPerspective
	maxRounds            int
	convergenceThreshold float64
	model                string
}

func NewDelphiEngine(spawner openclaw.Spawner, cc openclaw.ChatCompleter, llmModel string) *DelphiEngine {
	return &DelphiEngine{
		spawner:              spawner,
		cc:                   cc,
		mediator:             NewMediator(cc, llmModel),
		calibrator:           NewCalibrator(),
		experts:              DefaultExperts(),
		maxRounds:            DefaultMaxRounds,
		convergenceThreshold: DefaultConvergenceThreshold,
		model:                llmModel,
	}
}

func (d *DelphiEngine) SetExperts(experts []model.ExpertPerspective) { d.experts = experts }
func (d *DelphiEngine) SetMaxRounds(n int)                          { d.maxRounds = n }
func (d *DelphiEngine) SetCalibrator(c *Calibrator)                 { d.calibrator = c }

// Refine runs the Delphi process on a set of questions given simulation data.
func (d *DelphiEngine) Refine(ctx context.Context, simSummary string, questions []string) ([]model.PredictionResult, error) {
	results := make([]model.PredictionResult, len(questions))

	for qi, question := range questions {
		result, err := d.refineOne(ctx, simSummary, question)
		if err != nil {
			return nil, fmt.Errorf("question %d: %w", qi, err)
		}
		results[qi] = *result
	}
	return results, nil
}

func (d *DelphiEngine) refineOne(ctx context.Context, simSummary, question string) (*model.PredictionResult, error) {
	var prevFeedback string
	var expertPreds map[string]ExpertOutput
	var roundCount int

	for round := 1; round <= d.maxRounds; round++ {
		roundCount = round

		// Spawn all experts concurrently
		preds, err := d.spawnExperts(ctx, question, simSummary, prevFeedback, round)
		if err != nil {
			return nil, fmt.Errorf("round %d experts: %w", round, err)
		}
		expertPreds = preds

		// Check convergence
		probs := extractProbabilities(expertPreds)
		if StdDev(probs) < d.convergenceThreshold {
			break
		}

		// Mediate if not converged and not last round
		if round < d.maxRounds {
			medResult, err := d.mediator.Mediate(ctx, question, expertPreds)
			if err == nil {
				prevFeedback = medResult.Feedback
			}
		}
	}

	// Aggregate and calibrate
	probs := extractProbabilities(expertPreds)
	rawMedian := Median(probs)
	calibrated := d.calibrator.Calibrate(rawMedian)

	expertViews := make(map[string]float64)
	var keyArgs []string
	var disagreements []string
	for name, ep := range expertPreds {
		expertViews[name] = ep.Probability
		if ep.Reasoning != "" {
			keyArgs = append(keyArgs, fmt.Sprintf("[%s] %s", name, ep.Reasoning))
		}
	}

	// Get final disagreements from last mediation
	if prevFeedback != "" {
		disagreements = append(disagreements, prevFeedback)
	}

	return &model.PredictionResult{
		Question:      question,
		Probability:   calibrated,
		Confidence:    1.0 - StdDev(probs),
		KeyArguments:  keyArgs,
		Disagreements: disagreements,
		ExpertViews:   expertViews,
		Rounds:        roundCount,
	}, nil
}

func (d *DelphiEngine) spawnExperts(ctx context.Context, question, simSummary, feedback string, round int) (map[string]ExpertOutput, error) {
	type result struct {
		name   string
		output ExpertOutput
		err    error
	}

	ch := make(chan result, len(d.experts))
	var wg sync.WaitGroup

	for _, expert := range d.experts {
		wg.Add(1)
		go func(exp model.ExpertPerspective) {
			defer wg.Done()
			ep, err := d.spawnOneExpert(ctx, exp, question, simSummary, feedback, round)
			ch <- result{name: exp.Name, output: ep, err: err}
		}(expert)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	preds := make(map[string]ExpertOutput)
	for r := range ch {
		if r.err == nil {
			preds[r.name] = r.output
		}
	}

	if len(preds) == 0 {
		return nil, fmt.Errorf("all experts failed")
	}
	return preds, nil
}

func (d *DelphiEngine) spawnOneExpert(ctx context.Context, expert model.ExpertPerspective, question, simSummary, feedback string, round int) (ExpertOutput, error) {
	prompt := fmt.Sprintf(`%s

Question to predict: %s

Simulation data summary:
%s`, expert.SystemPrompt, question, simSummary)

	if feedback != "" {
		prompt += fmt.Sprintf("\n\nFeedback from previous round's mediator:\n%s\n\nPlease reconsider your prediction in light of this feedback.", feedback)
	}

	prompt += `

Respond with JSON: {"probability": 0.XX, "reasoning": "your key reasoning in 1-2 sentences"}`

	resp, err := d.spawner.Spawn(ctx, openclaw.SpawnRequest{
		Task:              prompt,
		Model:             d.model,
		Label:             fmt.Sprintf("delphi-r%d-%s", round, expert.Name),
		RunTimeoutSeconds: 60,
		Deliver:           true,
	})
	if err != nil {
		return ExpertOutput{}, err
	}

	subResult, err := d.spawner.WaitForResult(ctx, resp.RunID, time.Minute)
	if err != nil {
		return ExpertOutput{}, err
	}

	var ep ExpertOutput
	if err := json.Unmarshal([]byte(subResult.Output), &ep); err != nil {
		return ExpertOutput{Probability: 0.5, Reasoning: subResult.Output}, nil
	}
	return ep, nil
}

func extractProbabilities(preds map[string]ExpertOutput) []float64 {
	vals := make([]float64, 0, len(preds))
	for _, p := range preds {
		vals = append(vals, p.Probability)
	}
	return vals
}
