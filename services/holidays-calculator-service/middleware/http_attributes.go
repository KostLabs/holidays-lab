package middleware

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// HTTPSemanticConventions ensures key HTTP semantic convention attributes
// (http.route, http.request.method) are always present on the current span.
func HTTPSemanticConventions() gin.HandlerFunc {
	return func(c *gin.Context) {
		span := trace.SpanFromContext(c.Request.Context())
		if span == nil {
			c.Next()
			return
		}

		route := c.FullPath()
		if route == "" {
			route = c.Request.URL.Path
		}

		method := c.Request.Method

		span.SetAttributes(
			semconv.HTTPRoute(route),
			attribute.String(string(semconv.HTTPRequestMethodKey), method),
		)

		c.Next()
	}
}
