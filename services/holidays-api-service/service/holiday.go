package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"holidays-api-service/model"
	"holidays-api-service/repository"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type HolidayService interface {
	FetchHolidays(ctx context.Context, year string) (*model.HolidaysResponse, error)
}

type holidayService struct {
	repo        repository.HolidayRepository
	externalURL string
	httpClient  *http.Client
}

func NewHolidayService(repo repository.HolidayRepository, externalURL string) HolidayService {
	return &holidayService{
		repo:        repo,
		externalURL: externalURL,
		httpClient: &http.Client{
			Timeout:   10 * time.Second,
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		},
	}
}

func (s *holidayService) FetchHolidays(ctx context.Context, yearStr string) (*model.HolidaysResponse, error) {
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
	holidays, err := s.repo.FindByYear(dbCtx, year)
	dbSpan.End()
	if err != nil {
		log.Printf("Error querying database: %v", err)
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
		parsedDate, err := time.Parse("2006-01-02", h.Date)
		if err != nil {
			log.Printf("failed to parse holiday date %q: %v", h.Date, err)
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
		tracer := otel.Tracer("holidays-api-service/service")
		spanCtx, span := tracer.Start(parentCtx, "AsyncSaveHolidays", trace.WithSpanKind(trace.SpanKindInternal))
		defer span.End()

		if err := s.repo.SaveMany(spanCtx, holidays); err != nil {
			log.Printf("Failed to save holidays to database: %v", err)
		} else {
			log.Printf("Successfully saved %d holidays for year %d to database", len(holidays), year)
		}
	}(bgCtx, holidays, year)

	return &model.HolidaysResponse{
		Holidays: holidays,
		Source:   "api",
	}, nil
}

func (s *holidayService) fetchFromExternalAPI(ctx context.Context, year string) ([]model.Holiday, error) {
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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("external API returned status code: %d", resp.StatusCode)
	}

	var holidays []model.Holiday
	if err := json.NewDecoder(resp.Body).Decode(&holidays); err != nil {
		return nil, fmt.Errorf("failed to decode API response: %w", err)
	}

	return holidays, nil
}
