package api

import (
	"net/http"
	"time"

	"swarm-predict/internal/cognitive"
	"swarm-predict/internal/model"
	"swarm-predict/internal/simulation"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SimulationHandler struct {
	deps    Deps
	engines map[string]*simulation.Engine
	states  map[string]*model.SimulationState
}

func NewSimulationHandler(deps Deps) *SimulationHandler {
	return &SimulationHandler{
		deps:    deps,
		engines: make(map[string]*simulation.Engine),
		states:  make(map[string]*model.SimulationState),
	}
}

func (h *SimulationHandler) Create(c *gin.Context) {
	var req struct {
		ProjectID      string               `json:"project_id" binding:"required"`
		Rounds         int                  `json:"rounds"`
		AgentsPerGroup int                  `json:"agents_per_group"`
		Agents         []model.AgentProfile `json:"agents" binding:"required"`
		Events         []model.EventConfig  `json:"events"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Rounds <= 0 {
		req.Rounds = 40
	}
	if req.AgentsPerGroup <= 0 {
		req.AgentsPerGroup = 6
	}

	simID := uuid.New().String()
	config := model.SimulationConfig{
		ID:             simID,
		ProjectID:      req.ProjectID,
		Rounds:         req.Rounds,
		AgentsPerGroup: req.AgentsPerGroup,
		Events:         req.Events,
		Model:          h.deps.Model,
		FlashModel:     h.deps.FlashModel,
	}

	socialStore := cognitive.NewInMemorySocialStore()
	socialGraph := cognitive.NewSocialGraph(socialStore)
	engine := simulation.NewEngine(h.deps.Spawner, h.deps.ChatClient, config,
		simulation.WithSocialGraph(socialGraph),
	)
	h.engines[simID] = engine
	h.states[simID] = &model.SimulationState{
		ID:          simID,
		Status:      "pending",
		TotalRounds: req.Rounds,
	}

	c.JSON(http.StatusOK, gin.H{"simulation_id": simID, "status": "created"})
}

func (h *SimulationHandler) Start(c *gin.Context) {
	simID := c.Param("id")
	engine, ok := h.engines[simID]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "simulation not found"})
		return
	}

	var req struct {
		Agents []model.AgentProfile `json:"agents" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	state := h.states[simID]
	state.Status = "running"
	state.StartedAt = time.Now()

	go func() {
		_, err := engine.Run(c.Request.Context(), req.Agents)
		now := time.Now()
		if err != nil {
			state.Status = "failed"
			state.Error = err.Error()
		} else {
			state.Status = "completed"
		}
		state.CompletedAt = &now
	}()

	c.JSON(http.StatusOK, gin.H{"status": "started"})
}

func (h *SimulationHandler) Status(c *gin.Context) {
	simID := c.Param("id")
	state, ok := h.states[simID]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "simulation not found"})
		return
	}
	c.JSON(http.StatusOK, state)
}
