package service

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ProxyService interface {
	Forward(req ProxyRequest) (*ProxyResponse, error)
}

type proxyService struct {
	client *http.Client
}

func NewProxyService() ProxyService {
	return &proxyService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type ProxyRequest struct {
	Method  string
	URL     string
	Body    []byte
	Headers map[string]string
}

type ProxyResponse struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

func (s *proxyService) Forward(req ProxyRequest) (*ProxyResponse, error) {
	var bodyReader io.Reader
	if req.Body != nil {
		bodyReader = bytes.NewReader(req.Body)
	}

	httpReq, err := http.NewRequest(req.Method, req.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Copy headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	resp, err := s.client.Do(httpReq)
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
