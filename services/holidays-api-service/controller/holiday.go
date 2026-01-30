package controller

import (
	"net/http"

	"holidays-api-service/model"
	"holidays-api-service/service"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type HolidayController struct {
	holidayService service.HolidayService
}

func NewHolidayController(holidayService service.HolidayService) *HolidayController {
	return &HolidayController{holidayService: holidayService}
}

// FetchHolidays handles GET /fetch?year=XXXX
func (c *HolidayController) FetchHolidays(ctx *gin.Context) {
	tracer := otel.Tracer("holidays-api-service/controller")
	spanCtx, span := tracer.Start(ctx.Request.Context(), "HolidayController.FetchHolidays", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	var req model.FetchRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("holidays.year", ctx.Query("year")))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	span.SetAttributes(attribute.String("holidays.year", req.Year))

	resp, err := c.holidayService.FetchHolidays(spanCtx, req.Year)
	if err != nil {
		span.RecordError(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
