package api

import (
	"net/http"

	"swarm-predict/internal/prediction"

	"github.com/gin-gonic/gin"
)

type PredictionHandler struct {
	deps Deps
}

func NewPredictionHandler(deps Deps) *PredictionHandler {
	return &PredictionHandler{deps: deps}
}

func (h *PredictionHandler) Refine(c *gin.Context) {
	var req struct {
		SimSummary string   `json:"sim_summary" binding:"required"`
		Questions  []string `json:"questions" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	engine := prediction.NewDelphiEngine(h.deps.Spawner, h.deps.ChatClient, h.deps.Model)
	results, err := engine.Refine(c.Request.Context(), req.SimSummary, req.Questions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"predictions": results})
}
