package holidaysclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"holidays-calculator-service/model"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// HolidaysResponse represents the JSON structure returned by holidays-api-service.
type HolidaysResponse struct {
	Holidays []model.Holiday `json:"holidays"`
	Source   string          `json:"source"`
}

// Client is a simple HTTP client for holidays-api-service.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout:   10 * time.Second,
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		},
	}
}

// FetchHolidays fetches holidays for the given year from holidays-api-service.
func (c *Client) FetchHolidays(ctx context.Context, year string) (*HolidaysResponse, error) {
	url := fmt.Sprintf("%s/fetch?year=%s", c.baseURL, year)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("holidays-api-service returned status code: %d", resp.StatusCode)
	}

	var hr HolidaysResponse
	decodeErr := json.NewDecoder(resp.Body).Decode(&hr)
	if decodeErr != nil {
		return nil, fmt.Errorf("failed to decode holidays-api-service response: %w", decodeErr)
	}

	return &hr, nil
}
