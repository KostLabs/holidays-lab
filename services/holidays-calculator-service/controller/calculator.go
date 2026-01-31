package controller

import (
	"context"
	"net/http"

	"holidays-calculator-service/model"
	"holidays-calculator-service/service"

	"github.com/gin-gonic/gin"
)

type ICalculatorService interface {
	CalculateDaysUntil(ctx context.Context, fromDate, name string) (*model.CalculateResponse, error)
}

type CalculatorController struct {
	service ICalculatorService
}

func NewCalculatorController(service ICalculatorService) *CalculatorController {
	return &CalculatorController{service: service}
}

// Calculate handles GET /calculate?date=YYYY-MM-DD&name=holiday_name
func (c *CalculatorController) Calculate(ctx *gin.Context) {
	var req model.CalculateRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid query parameters",
			"details": err.Error(),
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
