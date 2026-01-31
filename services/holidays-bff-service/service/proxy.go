package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

type ProxyRequest struct {
	Method        string
	URL           string
	Body          []byte
	Headers       map[string]string
	TargetService string // logical downstream service name (e.g., holidays-calculator-service)
}

type ProxyResponse struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

type ProxyService struct {
	httpClient *http.Client
}

func NewProxyService() *ProxyService {
	return &ProxyService{
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *ProxyService) Forward(ctx context.Context, req ProxyRequest) (*ProxyResponse, error) {
	tracer := otel.Tracer("holidays-bff-service/proxy")

	var bodyReader io.Reader
	if req.Body != nil {
		bodyReader = bytes.NewReader(req.Body)
	}

	parsedURL, err := url.Parse(req.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	spanName := fmt.Sprintf("%s %s", req.Method, parsedURL.Path)
	ctx, span := tracer.Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		semconv.HTTPRequestMethodKey.String(req.Method),
		semconv.URLFull(parsedURL.String()),
		attribute.String("server.address", parsedURL.Hostname()),
	)
	if req.TargetService != "" {
		span.SetAttributes(attribute.String("peer.service", req.TargetService))
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Copy headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Inject trace context for downstream services.
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(httpReq.Header))

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to forward request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return &ProxyResponse{
		StatusCode: resp.StatusCode,
		Body:       body,
		Headers:    resp.Header,
	}, nil
}
