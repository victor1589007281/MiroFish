package api

import (
	"swarm-predict/internal/graph"
	"swarm-predict/internal/openclaw"
	"swarm-predict/internal/store"

	"github.com/gin-gonic/gin"
)

// Deps holds all injected dependencies for the API layer.
type Deps struct {
	ChatClient openclaw.ChatCompleter
	Spawner    openclaw.Spawner
	Graph      graph.Client
	Store      store.ProjectStore
	Model      string
	FlashModel string
}

// NewRouter creates the Gin router with all API routes.
func NewRouter(deps Deps) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	api := r.Group("/api")
	{
		g := api.Group("/graph")
		{
			gh := NewGraphHandler(deps)
			g.POST("/ontology/generate", gh.GenerateOntology)
			g.POST("/build", gh.BuildGraph)
		}

		s := api.Group("/simulation")
		{
			sh := NewSimulationHandler(deps)
			s.POST("/create", sh.Create)
			s.POST("/:id/start", sh.Start)
			s.GET("/:id/status", sh.Status)
		}

		p := api.Group("/prediction")
		{
			ph := NewPredictionHandler(deps)
			p.POST("/:sim_id/refine", ph.Refine)
		}

		rpt := api.Group("/report")
		{
			rh := NewReportHandler(deps)
			rpt.POST("/generate", rh.Generate)
			rpt.GET("/:id", rh.Get)
			rpt.POST("/:id/chat", rh.Chat)
		}
	}

	// Full pipeline endpoint
	oh := NewOrchestratorHandler(deps)
	api.POST("/pipeline/run", oh.RunPipeline)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return r
}
