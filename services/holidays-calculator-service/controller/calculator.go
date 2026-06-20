package controller

import (
	"context"
	"net/http"

	"holidays-calculator-service/model"
	"holidays-calculator-service/service"

	"github.com/gin-gonic/gin"
)

type CalculatorService interface {
	CalculateDaysUntil(ctx context.Context, fromDate, name string) (*model.CalculateResponse, error)
}

type CalculatorController struct {
	service CalculatorService
}

func NewCalculatorController(service CalculatorService) *CalculatorController {
	return &CalculatorController{service: service}
}

// Calculate handles GET /calculate?date=YYYY-MM-DD&name=holiday_name
//goverifier:ignore:context-propagation
func (c *CalculatorController) Calculate(ctx *gin.Context) {
	var req model.CalculateRequest
	bindErr := ctx.ShouldBindQuery(&req)
	if bindErr != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid query parameters",
			"details": bindErr.Error(),
		})
		return
	}

	result, err := c.service.CalculateDaysUntil(ctx.Request.Context(), req.Date, req.Name)
	if err != nil {
		if err == service.ErrHolidayNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "holiday not found",
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to calculate days until holiday",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, result)
}
