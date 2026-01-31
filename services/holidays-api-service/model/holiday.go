package model

import "time"

// Holiday represents a holiday record as returned by the external API and stored in MongoDB.
type Holiday struct {
	Date      string    `json:"date" bson:"date"`
	EndDate   *string   `json:"end_date,omitempty" bson:"end_date,omitempty"`
	Title     string    `json:"title" bson:"title"`
	Notes     *string   `json:"notes,omitempty" bson:"notes,omitempty"`
	Kind      string    `json:"kind" bson:"kind"`
	KindID    string    `json:"kind_id" bson:"kind_id"`
	Year      int       `json:"year" bson:"year"`
	CreatedAt time.Time `json:"created_at,omitzero" bson:"created_at,omitempty"`
}

// FetchRequest represents query parameters for fetching holidays
type FetchRequest struct {
	Year string `form:"year" binding:"required"`
}

// HolidaysResponse represents the API response
type HolidaysResponse struct {
	Holidays []Holiday `json:"holidays"`
	Source   string    `json:"source"`
}
