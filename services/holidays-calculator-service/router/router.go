package router

import (
	"holidays-calculator-service/config"
	"holidays-calculator-service/controller"
	"holidays-calculator-service/middleware"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

type Router struct {
	engine *gin.Engine
	cfg    *config.Config
}

func NewRouter(cfg *config.Config) *Router {
	return &Router{
		engine: gin.Default(),
		cfg:    cfg,
	}
}

func (r *Router) Setup(calcController *controller.CalculatorController) {
	// HTTP server tracing
	r.engine.Use(otelgin.Middleware("holidays-calculator-service"))
	r.engine.Use(middleware.HTTPSemanticConventions())

	api := r.engine.Group(r.cfg.BasePath)
	{
		// GET /api/v1/holidays-calculator/calculate?date=YYYY-MM-DD&name=holiday_name
		api.GET("/calculate", calcController.Calculate)
	}

	r.engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})
}

func (r *Router) Engine() *gin.Engine {
	return r.engine
}
