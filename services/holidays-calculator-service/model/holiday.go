package model

import "time"

// Holiday mirrors the holiday representation returned by holidays-api-service.
type Holiday struct {
	Date      string    `json:"date"`
	EndDate   *string   `json:"end_date,omitempty"`
	Title     string    `json:"title"`
	Notes     *string   `json:"notes,omitempty"`
	Kind      string    `json:"kind"`
	KindID    string    `json:"kind_id"`
	Year      int       `json:"year"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}
