package simulation

import (
	"math"
	"strings"
)

// EmergenceSignal summarizes detected emergent patterns in a round.
type EmergenceSignal struct {
	PolarizationIndex float64  `json:"polarization_index"`
	TopTopics         []string `json:"top_topics"`
	InfluencerShifts  []string `json:"influencer_shifts"`
	ConsensusLevel    float64  `json:"consensus_level"`
}

// DetectEmergence analyzes action logs to find emergent patterns.
func DetectEmergence(logs []ActionLog) EmergenceSignal {
	if len(logs) == 0 {
		return EmergenceSignal{}
	}

	topicCounts := make(map[string]int)
	sentimentByAgent := make(map[string][]float64)

	for _, log := range logs {
		if log.Decision.ActionType == "post" || log.Decision.ActionType == "reply" {
			words := strings.Fields(strings.ToLower(log.Decision.Content))
			for _, w := range words {
				if len(w) > 3 {
					topicCounts[w]++
				}
			}
			sentimentByAgent[log.Decision.AgentID] = append(
				sentimentByAgent[log.Decision.AgentID],
				simpleSentiment(log.Decision.Content),
			)
		}
	}

	// Top topics
	var topTopics []string
	for topic, count := range topicCounts {
		if count >= 3 {
			topTopics = append(topTopics, topic)
		}
	}
	if len(topTopics) > 5 {
		topTopics = topTopics[:5]
	}

	// Polarization: variance of agent-level average sentiments
	var sentiments []float64
	for _, ss := range sentimentByAgent {
		avg := mean(ss)
		sentiments = append(sentiments, avg)
	}
	polarization := variance(sentiments)

	// Consensus: 1 - polarization (clamped)
	consensus := math.Max(0, 1-polarization*10)

	return EmergenceSignal{
		PolarizationIndex: polarization,
		TopTopics:         topTopics,
		ConsensusLevel:    consensus,
	}
}

func simpleSentiment(text string) float64 {
	lower := strings.ToLower(text)
	positive := strings.Count(lower, "good") + strings.Count(lower, "great") +
		strings.Count(lower, "agree") + strings.Count(lower, "support") +
		strings.Count(lower, "positive") + strings.Count(lower, "excellent")
	negative := strings.Count(lower, "bad") + strings.Count(lower, "terrible") +
		strings.Count(lower, "disagree") + strings.Count(lower, "oppose") +
		strings.Count(lower, "negative") + strings.Count(lower, "wrong")
	total := positive + negative
	if total == 0 {
		return 0
	}
	return float64(positive-negative) / float64(total)
}

func mean(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	var s float64
	for _, v := range vals {
		s += v
	}
	return s / float64(len(vals))
}

func variance(vals []float64) float64 {
	if len(vals) < 2 {
		return 0
	}
	m := mean(vals)
	var ss float64
	for _, v := range vals {
		d := v - m
		ss += d * d
	}
	return ss / float64(len(vals))
}
