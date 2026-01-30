package router

import (
	"holidays-api-service/config"
	"holidays-api-service/controller"

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

func (r *Router) Setup(holidayController *controller.HolidayController) {
	api := r.engine.Group(r.cfg.BasePath)
	{
		// GET /api/v1/holidays-proceeding/fetch?year=XXXX
		api.GET("/fetch", holidayController.FetchHolidays)
	}

	r.engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})
}

func (r *Router) Engine() *gin.Engine {
	return r.engine
}
