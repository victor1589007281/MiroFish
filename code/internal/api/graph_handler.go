package api

import (
	"net/http"

	"swarm-predict/internal/cognitive"
	"swarm-predict/internal/model"
	"swarm-predict/internal/openclaw"

	"github.com/gin-gonic/gin"
)

type GraphHandler struct {
	deps Deps
}

func NewGraphHandler(deps Deps) *GraphHandler {
	return &GraphHandler{deps: deps}
}

func (h *GraphHandler) GenerateOntology(c *gin.Context) {
	var req struct {
		Document    string `json:"document" binding:"required"`
		Description string `json:"description" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.deps.ChatClient.Complete(c.Request.Context(), openclaw.ChatRequest{
		Model: h.deps.Model,
		Messages: []openclaw.ChatMessage{
			{Role: "system", Content: ontologySystemPrompt},
			{Role: "user", Content: "Document:\n" + req.Document + "\n\nPrediction goal: " + req.Description},
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var ontology model.OntologyResult
	if err := cognitive.ExtractJSON(resp, &ontology); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse ontology", "raw": cognitive.ExtractContent(resp)})
		return
	}

	c.JSON(http.StatusOK, ontology)
}

func (h *GraphHandler) BuildGraph(c *gin.Context) {
	var req struct {
		Ontology model.OntologyResult `json:"ontology" binding:"required"`
		Document string               `json:"document" binding:"required"`
		GraphID  string               `json:"graph_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for _, et := range req.Ontology.EntityTypes {
		h.deps.Graph.AddEntity(c.Request.Context(), et.Name, et.Description, nil)
	}
	for _, rt := range req.Ontology.RelationTypes {
		h.deps.Graph.AddRelation(c.Request.Context(), rt.Source, rt.Target, rt.Name, map[string]interface{}{
			"description": rt.Description,
		})
	}

	c.JSON(http.StatusOK, gin.H{"status": "built", "graph_id": req.GraphID})
}

const ontologySystemPrompt = `Analyze the provided document and prediction goal. Extract entity types and relation types.
Output JSON: {"entity_types": [{"name": "...", "description": "..."}], "relation_types": [{"name": "...", "source": "entity_type", "target": "entity_type", "description": "..."}]}`
