package controller

import (
	"net/http"

	"holidays-api-service/model"
	"holidays-api-service/service"

	"github.com/gin-gonic/gin"
)

type HolidayController struct {
	holidayService service.HolidayService
}

func NewHolidayController(holidayService service.HolidayService) *HolidayController {
	return &HolidayController{holidayService: holidayService}
}

// FetchHolidays handles GET /fetch?year=XXXX
func (c *HolidayController) FetchHolidays(ctx *gin.Context) {
	var req model.FetchRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := c.holidayService.FetchHolidays(ctx.Request.Context(), req.Year)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
