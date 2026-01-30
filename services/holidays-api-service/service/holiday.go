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
			Timeout: 10 * time.Second,
		},
	}
}

func (s *holidayService) FetchHolidays(ctx context.Context, yearStr string) (*model.HolidaysResponse, error) {
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		return nil, fmt.Errorf("invalid year format: %w", err)
	}

	holidays, err := s.repo.FindByYear(ctx, year)
	if err != nil {
		log.Printf("Error querying database: %v", err)
	}

	if len(holidays) > 0 {
		return &model.HolidaysResponse{
			Holidays: holidays,
			Source:   "database",
		}, nil
	}

	holidays, err = s.fetchFromExternalAPI(yearStr)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from external API: %w", err)
	}

	// Filter holidays to only those matching the requested year based on the date
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

	go func() {
		bgCtx := context.Background()
		if err := s.repo.SaveMany(bgCtx, holidays); err != nil {
			log.Printf("Failed to save holidays to database: %v", err)
		} else {
			log.Printf("Successfully saved %d holidays for year %d to database", len(holidays), year)
		}
	}()

	return &model.HolidaysResponse{
		Holidays: holidays,
		Source:   "api",
	}, nil
}

func (s *holidayService) fetchFromExternalAPI(year string) ([]model.Holiday, error) {
	url := fmt.Sprintf("%s&year=%s", s.externalURL, year)

	resp, err := s.httpClient.Get(url)
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
