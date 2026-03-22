package simulation

import "swarm-predict/internal/model"

// GroupAgents splits agents into groups of the given size.
func GroupAgents(agents []model.AgentProfile, groupSize int) [][]model.AgentProfile {
	if groupSize <= 0 {
		groupSize = 6
	}
	var groups [][]model.AgentProfile
	for i := 0; i < len(agents); i += groupSize {
		end := i + groupSize
		if end > len(agents) {
			end = len(agents)
		}
		groups = append(groups, agents[i:end])
	}
	return groups
}
