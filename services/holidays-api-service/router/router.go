package router

import (
	"holidays-api-service/config"
	"holidays-api-service/controller"
	"holidays-api-service/middleware"

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

func (r *Router) Setup(holidayController *controller.HolidayController) {
	// HTTP server tracing
	r.engine.Use(otelgin.Middleware("holidays-api-service"))
	r.engine.Use(middleware.HTTPSemanticConventions())

	api := r.engine.Group(r.cfg.BasePath)
	{
		// GET /api/v1/holidays-proceeding/fetch?year=XXXX
		api.GET("/fetch", holidayController.FetchHolidays)
	}
}

func (r *Router) Engine() *gin.Engine {
	return r.engine
}
