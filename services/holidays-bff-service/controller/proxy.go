package controller

import (
	"fmt"
	"io"
	"net/http"

	"holidays-bff-service/config"
	"holidays-bff-service/model"
	"holidays-bff-service/service"

	"github.com/gin-gonic/gin"
)

type HolidaysController struct {
	config       *config.Config
	proxyService service.ProxyService
}

func NewHolidaysController(cfg *config.Config, proxyService service.ProxyService) *HolidaysController {
	return &HolidaysController{
		config:       cfg,
		proxyService: proxyService,
	}
}

// GetHolidaysByYear handles GET /holidays?year=XXXX
// Binds query params to FetchHolidaysRequest model
// Forwards to holidays_api_service + /fetch
func (c *HolidaysController) GetHolidaysByYear(ctx *gin.Context) {
	var req model.FetchHolidaysRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.forwardToService(ctx, "holidays_api_service", "/fetch")
}

// GetHolidaysByDateAndName handles GET /holidays?date=YYYY-MM-DD&name=XXXX
// Binds query params to CalculateHolidaysRequest model
// Forwards to holidays_calculator_service + /calculate
func (c *HolidaysController) GetHolidaysByDateAndName(ctx *gin.Context) {
	var req model.CalculateHolidaysRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.forwardToService(ctx, "holidays_calculator_service", "/calculate")
}

func (c *HolidaysController) forwardToService(ctx *gin.Context, serviceName string, endpoint string) {
	// Find the external service configuration
	external, found := c.config.GetExternalByName(serviceName)
	if !found {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("service %s not found", serviceName),
		})
		return
	}

	// Build target URL with the specific endpoint
	targetURL := fmt.Sprintf("%s%s", external.URL, endpoint)

	// Add query parameters
	if ctx.Request.URL.RawQuery != "" {
		targetURL = fmt.Sprintf("%s?%s", targetURL, ctx.Request.URL.RawQuery)
	}

	// Read request body
	var body []byte
	var err error
	if ctx.Request.Body != nil {
		body, err = io.ReadAll(ctx.Request.Body)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "failed to read request body",
			})
			return
		}
	}

	// Copy headers
	headers := make(map[string]string)
	for key, values := range ctx.Request.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	// Forward the request
	proxyReq := service.ProxyRequest{
		Method:        ctx.Request.Method,
		URL:           targetURL,
		Body:          body,
		Headers:       headers,
		TargetService: external.Name,
	}

	resp, err := c.proxyService.Forward(ctx.Request.Context(), proxyReq)
	if err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{
			"error": fmt.Sprintf("failed to forward request: %v", err),
		})
		return
	}

	// Copy response headers
	for key, values := range resp.Headers {
		for _, value := range values {
			ctx.Header(key, value)
		}
	}

	// Return the response
	ctx.Data(resp.StatusCode, resp.Headers.Get("Content-Type"), resp.Body)
}
