package react

import (
	"strings"
	"testing"

	"swarm-predict/internal/model"

	"github.com/stretchr/testify/assert"
)

func TestFormatMarkdown(t *testing.T) {
	report := &model.Report{
		Sections: []model.ReportSection{
			{Title: "Executive Summary", Content: "This is the summary.", Order: 1},
			{Title: "Analysis", Content: "Detailed analysis here.", Order: 2},
		},
		Predictions: []model.PredictionResult{
			{
				Question:    "Will X happen?",
				Probability: 0.75,
				Confidence:  0.85,
				Rounds:      3,
				ExpertViews: map[string]float64{"optimist": 0.8, "pessimist": 0.6},
				KeyArguments: []string{"[optimist] Strong economic signals"},
				Disagreements: []string{"Timing is uncertain"},
			},
		},
	}

	md := FormatMarkdown(report)
	assert.Contains(t, md, "# 预测分析报告")
	assert.Contains(t, md, "## Executive Summary")
	assert.Contains(t, md, "## Analysis")
	assert.Contains(t, md, "Will X happen?")
	assert.Contains(t, md, "75%")
	assert.Contains(t, md, "85%")
	assert.Contains(t, md, "optimist")
	assert.Contains(t, md, "Strong economic signals")
	assert.Contains(t, md, "Timing is uncertain")
}

func TestFormatPlainText(t *testing.T) {
	report := &model.Report{
		Sections: []model.ReportSection{
			{Title: "Summary", Content: "Short summary.", Order: 1},
		},
		Predictions: []model.PredictionResult{
			{Question: "Q?", Probability: 0.6, Confidence: 0.7},
		},
	}

	text := FormatPlainText(report)
	assert.Contains(t, text, "[Summary]")
	assert.Contains(t, text, "60%")
}

func TestFormatMarkdown_NoPredictions(t *testing.T) {
	report := &model.Report{
		Sections: []model.ReportSection{
			{Title: "Test", Content: "Content.", Order: 1},
		},
	}

	md := FormatMarkdown(report)
	assert.Contains(t, md, "## Test")
	assert.False(t, strings.Contains(md, "预测结果汇总"))
}

func TestSummarizeSimulation(t *testing.T) {
	summary := SummarizeSimulation(40, 150, nil, nil)
	assert.Contains(t, summary, "40 rounds")
	assert.Contains(t, summary, "150 posts")
}
