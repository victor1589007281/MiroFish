package api

import (
	"net/http"

	"swarm-predict/internal/orchestrate"

	"github.com/gin-gonic/gin"
)

type OrchestratorHandler struct {
	deps Deps
}

func NewOrchestratorHandler(deps Deps) *OrchestratorHandler {
	return &OrchestratorHandler{deps: deps}
}

// RunPipeline executes the full six-step MiroFish V2 pipeline.
func (h *OrchestratorHandler) RunPipeline(c *gin.Context) {
	var req struct {
		Topic       string   `json:"topic" binding:"required"`
		Questions   []string `json:"questions" binding:"required"`
		AgentsCount int      `json:"agents_count"`
		SimRounds   int      `json:"sim_rounds"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.AgentsCount <= 0 {
		req.AgentsCount = 30
	}
	if req.SimRounds <= 0 {
		req.SimRounds = 40
	}

	config := orchestrate.Config{
		Model:       h.deps.Model,
		FlashModel:  h.deps.FlashModel,
		SimRounds:   req.SimRounds,
		AgentsCount: req.AgentsCount,
		GroupSize:   6,
	}

	pipeline := orchestrate.NewPipeline(h.deps.ChatClient, h.deps.Spawner, config)
	result, err := pipeline.Run(c.Request.Context(), req.Topic, req.Questions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
