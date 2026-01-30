package router

import (
	"holidays-bff-service/config"
	"holidays-bff-service/controller"
	"holidays-bff-service/service"

	"github.com/gin-gonic/gin"
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

func (r *Router) Setup() {
	proxyService := service.NewProxyService()
	holidaysController := controller.NewHolidaysController(r.cfg, proxyService)

	r.setupRoutes(holidaysController)

	// Health check endpoint
	r.engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
		})
	})
}

func (r *Router) setupRoutes(holidaysController *controller.HolidaysController) {
	// Register at base path
	v1 := r.engine.Group(r.cfg.BasePath)
	{
		// GET /api/v1/holidays-bff/holidays?year=XXXX
		v1.GET("/holidays", holidaysController.GetHolidaysByYear)

		// GET /api/v1/holidays-bff/holidays/calculate?date=YYYY-MM-DD&name=XXXX
		v1.GET("/holidays/calculate", holidaysController.GetHolidaysByDateAndName)
	}
}

func (r *Router) Engine() *gin.Engine {
	return r.engine
}
