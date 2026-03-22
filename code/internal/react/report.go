package react

import (
	"fmt"
	"strings"

	"swarm-predict/internal/model"
)

// FormatMarkdown converts a Report into a markdown string.
func FormatMarkdown(report *model.Report) string {
	var sb strings.Builder

	sb.WriteString("# 预测分析报告\n\n")
	sb.WriteString("---\n\n")

	for _, section := range report.Sections {
		fmt.Fprintf(&sb, "## %s\n\n%s\n\n", section.Title, section.Content)
	}

	if len(report.Predictions) > 0 {
		sb.WriteString("---\n\n## 预测结果汇总\n\n")
		sb.WriteString("| 问题 | 概率 | 置信度 | 迭代轮数 |\n")
		sb.WriteString("|------|------|--------|----------|\n")
		for _, p := range report.Predictions {
			fmt.Fprintf(&sb, "| %s | %.0f%% | %.0f%% | %d |\n",
				p.Question, p.Probability*100, p.Confidence*100, p.Rounds)
		}
		sb.WriteString("\n")

		for _, p := range report.Predictions {
			fmt.Fprintf(&sb, "### %s\n\n", p.Question)
			fmt.Fprintf(&sb, "**概率**: %.1f%%　**置信度**: %.1f%%\n\n", p.Probability*100, p.Confidence*100)

			if len(p.ExpertViews) > 0 {
				sb.WriteString("**专家视角**:\n\n")
				for name, prob := range p.ExpertViews {
					fmt.Fprintf(&sb, "- %s: %.0f%%\n", name, prob*100)
				}
				sb.WriteString("\n")
			}

			if len(p.KeyArguments) > 0 {
				sb.WriteString("**核心论据**:\n\n")
				for _, arg := range p.KeyArguments {
					fmt.Fprintf(&sb, "- %s\n", arg)
				}
				sb.WriteString("\n")
			}

			if len(p.Disagreements) > 0 {
				sb.WriteString("**分歧点**:\n\n")
				for _, d := range p.Disagreements {
					fmt.Fprintf(&sb, "- %s\n", d)
				}
				sb.WriteString("\n")
			}
		}
	}

	return sb.String()
}

// FormatPlainText renders a concise text summary of the report.
func FormatPlainText(report *model.Report) string {
	var sb strings.Builder

	for _, section := range report.Sections {
		fmt.Fprintf(&sb, "[%s]\n%s\n\n", section.Title, section.Content)
	}

	for _, p := range report.Predictions {
		fmt.Fprintf(&sb, "Q: %s → P=%.0f%% (confidence=%.0f%%)\n", p.Question, p.Probability*100, p.Confidence*100)
	}

	return sb.String()
}

// SummarizeSimulation creates a text summary of simulation results for LLM context.
func SummarizeSimulation(totalRounds, totalPosts int, logs interface{}, emergence interface{}) string {
	return fmt.Sprintf("Simulation completed: %d rounds, %d posts generated.\nEmergence signals and action logs available via tools.",
		totalRounds, totalPosts)
}
