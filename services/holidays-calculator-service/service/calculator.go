package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"holidays-calculator-service/model"
	"holidays-calculator-service/pkg/holidaysclient"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var ErrHolidayNotFound = errors.New("holiday not found")

type CalculatorService interface {
	CalculateDaysUntil(ctx context.Context, fromDate, name string) (*model.CalculateResponse, error)
}

type calculatorService struct {
	client *holidaysclient.Client
}

func NewCalculatorService(client *holidaysclient.Client) CalculatorService {
	return &calculatorService{client: client}
}

func (s *calculatorService) CalculateDaysUntil(ctx context.Context, fromDate, name string) (*model.CalculateResponse, error) {
	tracer := otel.Tracer("holidays-calculator-service/service")
	ctx, span := tracer.Start(ctx, "CalculateDaysUntil", trace.WithSpanKind(trace.SpanKindInternal))
	span.SetAttributes(
		attribute.String("holidays.from_date", fromDate),
		attribute.String("holidays.name", name),
	)
	defer span.End()

	parsedFrom, err := time.Parse("2006-01-02", fromDate)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	year := parsedFrom.Year()

	// Try in the same year first, then next year if needed.
	holiday, source, err := s.findHolidayInYear(ctx, year, name, parsedFrom)
	if err == ErrHolidayNotFound {
		holiday, source, err = s.findHolidayInYear(ctx, year+1, name, parsedFrom)
	}
	if err != nil {
		return nil, err
	}

	holidayDate, err := time.Parse("2006-01-02", holiday.Date)
	if err != nil {
		return nil, fmt.Errorf("failed to parse holiday date: %w", err)
	}

	if holidayDate.Before(parsedFrom) {
		return nil, fmt.Errorf("computed holiday date is before from date")
	}

	daysLeft := int(holidayDate.Sub(parsedFrom).Hours() / 24)

	return &model.CalculateResponse{
		HolidayName: holiday.Title,
		HolidayDate: holidayDate.Format("2006-01-02"),
		FromDate:    parsedFrom.Format("2006-01-02"),
		DaysLeft:    daysLeft,
		Source:      source,
	}, nil
}

func (s *calculatorService) findHolidayInYear(ctx context.Context, year int, name string, fromDate time.Time) (*model.Holiday, string, error) {
	tracer := otel.Tracer("holidays-calculator-service/service")
	ctx, span := tracer.Start(ctx, "FindHolidayInYear", trace.WithSpanKind(trace.SpanKindInternal))
	span.SetAttributes(
		attribute.Int("holidays.year", year),
	)
	defer span.End()

	resp, err := s.client.FetchHolidays(ctx, fmt.Sprintf("%d", year))
	if err != nil {
		return nil, "", err
	}

	needle := strings.ToLower(strings.TrimSpace(name))
	var best *model.Holiday

	for i := range resp.Holidays {
		h := &resp.Holidays[i]
		titleLower := strings.ToLower(strings.TrimSpace(h.Title))

		// Match by case-insensitive equality or substring.
		if titleLower == needle || strings.Contains(titleLower, needle) {
			// Parse date to ensure it is on or after fromDate when same year.
			parsedDate, err := time.Parse("2006-01-02", h.Date)
			if err != nil {
				continue
			}

			// For the same year, require the holiday to be on/after fromDate.
			if parsedDate.Year() == fromDate.Year() && parsedDate.Before(fromDate) {
				continue
			}

			if best == nil || parsedDate.Before(parseDateMust(best.Date)) {
				best = h
			}
		}
	}

	if best == nil {
		return nil, "", ErrHolidayNotFound
	}

	return best, resp.Source, nil
}

func parseDateMust(dateStr string) time.Time {
	t, _ := time.Parse("2006-01-02", dateStr)
	return t
}
