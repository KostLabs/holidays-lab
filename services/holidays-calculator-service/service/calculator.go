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

type CalculatorService struct {
	holidayClient *holidaysclient.Client
}

func NewCalculatorService(holidayClient *holidaysclient.Client) *CalculatorService {
	return &CalculatorService{holidayClient: holidayClient}
}

func (s *CalculatorService) CalculateDaysUntil(ctx context.Context, fromDate, name string) (*model.CalculateResponse, error) {
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

func (s *CalculatorService) findHolidayInYear(ctx context.Context, year int, name string, fromDate time.Time) (*model.Holiday, string, error) {
	tracer := otel.Tracer("holidays-calculator-service/service")
	ctx, span := tracer.Start(ctx, "FindHolidayInYear", trace.WithSpanKind(trace.SpanKindInternal))
	span.SetAttributes(attribute.Int("holidays.year", year))
	defer span.End()

	response, err := s.holidayClient.FetchHolidays(ctx, fmt.Sprintf("%d", year))
	if err != nil {
		return nil, "", err
	}

	needle := strings.ToLower(strings.TrimSpace(name))
	var foundHoliday *model.Holiday

	for i := range response.Holidays {
		h := &response.Holidays[i]
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

			if foundHoliday == nil || parsedDate.Before(parseDate(foundHoliday.Date)) {
				foundHoliday = h
			}
		}
	}

	if foundHoliday == nil {
		return nil, "", ErrHolidayNotFound
	}

	return foundHoliday, response.Source, nil
}

func parseDate(date string) time.Time {
	parsedDate, _ := time.Parse("2006-01-02", date)
	return parsedDate
}
