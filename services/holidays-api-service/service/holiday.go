package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"holidays-api-service/model"

	"github.com/KostLabs/golog"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type HolidayRepository interface {
	FindByYear(ctx context.Context, year int) ([]model.Holiday, error)
	SaveMany(ctx context.Context, holidays []model.Holiday) error
}

type HolidayService struct {
	repository  HolidayRepository
	externalURL string
	httpClient  *http.Client
}

func NewHolidayService(repo HolidayRepository, externalURL string) *HolidayService {
	return &HolidayService{
		repository:  repo,
		externalURL: externalURL,
		httpClient: &http.Client{
			Timeout:   10 * time.Second,
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		},
	}
}

func (s *HolidayService) FetchHolidays(ctx context.Context, yearStr string) (*model.HolidaysResponse, error) {
	tracer := otel.Tracer("holidays-api-service/service")
	ctx, span := tracer.Start(ctx, "FetchHolidaysPipeline", trace.WithSpanKind(trace.SpanKindInternal))
	span.SetAttributes(attribute.String("holidays.year", yearStr))
	defer span.End()

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		return nil, fmt.Errorf("invalid year format: %w", err)
	}

	// DB-first lookup span
	dbCtx, dbSpan := tracer.Start(ctx, "FetchHolidays.DBLookup", trace.WithSpanKind(trace.SpanKindInternal))
	holidays, err := s.repository.FindByYear(dbCtx, year)
	dbSpan.End()
	if err != nil {
		golog.Error("DatabaseError: querying holidays", map[string]any{"year": year, "error": err.Error()})
	}

	if len(holidays) > 0 {
		span.SetAttributes(attribute.Bool("holidays.cache_hit", true))
		return &model.HolidaysResponse{
			Holidays: holidays,
			Source:   "database",
		}, nil
	}

	span.SetAttributes(attribute.Bool("holidays.cache_hit", false))
	holidays, err = s.fetchFromExternalAPI(ctx, yearStr)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from external API: %w", err)
	}

	// Filter holidays to only those matching the requested year based on the date
	_, filterSpan := tracer.Start(ctx, "FetchHolidays.FilterByYear", trace.WithSpanKind(trace.SpanKindInternal))
	defer filterSpan.End()

	filtered := make([]model.Holiday, 0, len(holidays))
	for _, h := range holidays {
		parsedDate, parseErr := time.Parse("2006-01-02", h.Date)
		if parseErr != nil {
			golog.Error("ValidationError: failed to parse holiday date", map[string]any{"date": h.Date, "error": parseErr.Error()})
			continue
		}

		if parsedDate.Year() == year {
			h.Year = year
			filtered = append(filtered, h)
		}
	}
	holidays = filtered

	// Preserve the parent trace context for the async save, but
	// detach it from HTTP cancellation so it can complete even
	// after the request finishes.
	bgCtx := context.WithoutCancel(ctx)
	go func(parentCtx context.Context, holidays []model.Holiday, year int) {
		asyncTracer := otel.Tracer("holidays-api-service/service")
		spanCtx, asyncSpan := asyncTracer.Start(parentCtx, "AsyncSaveHolidays", trace.WithSpanKind(trace.SpanKindInternal))
		defer asyncSpan.End()

		saveErr := s.repository.SaveMany(spanCtx, holidays)
		if saveErr != nil {
			golog.Error("DatabaseError: failed to save holidays", map[string]any{"year": year, "error": saveErr.Error()})
		} else {
			golog.Info("successfully saved holidays to database", map[string]any{"year": year, "count": len(holidays)})
		}
	}(bgCtx, holidays, year)

	return &model.HolidaysResponse{
		Holidays: holidays,
		Source:   "api",
	}, nil
}

func (s *HolidayService) fetchFromExternalAPI(ctx context.Context, year string) ([]model.Holiday, error) {
	tracer := otel.Tracer("holidays-api-service/service")
	ctx, span := tracer.Start(ctx, "FetchFromExternalAPI", trace.WithSpanKind(trace.SpanKindInternal))
	span.SetAttributes(attribute.String("holidays.external.year", year))
	defer span.End()

	url := fmt.Sprintf("%s&year=%s", s.externalURL, year)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("external API returned status code: %d", resp.StatusCode)
	}

	var holidays []model.Holiday
	decodeErr := json.NewDecoder(resp.Body).Decode(&holidays)
	if decodeErr != nil {
		return nil, fmt.Errorf("failed to decode API response: %w", decodeErr)
	}

	return holidays, nil
}
