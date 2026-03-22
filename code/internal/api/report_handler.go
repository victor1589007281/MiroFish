package api

import (
	"net/http"
	"time"

	"swarm-predict/internal/model"
	"swarm-predict/internal/react"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ReportHandler struct {
	deps    Deps
	reports map[string]*model.Report
}

func NewReportHandler(deps Deps) *ReportHandler {
	return &ReportHandler{
		deps:    deps,
		reports: make(map[string]*model.Report),
	}
}

func (h *ReportHandler) Generate(c *gin.Context) {
	var req struct {
		ProjectDesc string                 `json:"project_desc" binding:"required"`
		SimSummary  string                 `json:"sim_summary" binding:"required"`
		Predictions []model.PredictionResult `json:"predictions"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	engine := react.NewEngine(h.deps.ChatClient, h.deps.Model)
	report, err := engine.GenerateReport(c.Request.Context(), req.ProjectDesc, req.SimSummary, req.Predictions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	report.ID = uuid.New().String()
	report.CreatedAt = time.Now()
	h.reports[report.ID] = report

	c.JSON(http.StatusOK, report)
}

func (h *ReportHandler) Get(c *gin.Context) {
	id := c.Param("id")
	report, ok := h.reports[id]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "report not found"})
		return
	}
	c.JSON(http.StatusOK, report)
}

func (h *ReportHandler) Chat(c *gin.Context) {
	id := c.Param("id")
	report, ok := h.reports[id]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "report not found"})
		return
	}

	var req struct {
		Question string `json:"question" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	engine := react.NewEngine(h.deps.ChatClient, h.deps.Model)
	answer, err := engine.Chat(c.Request.Context(), report, req.Question)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"answer": answer})
}
