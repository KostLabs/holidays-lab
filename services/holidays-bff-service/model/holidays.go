package model

// FetchHolidaysRequest represents query parameters for fetching holidays by year
type FetchHolidaysRequest struct {
	Year string `form:"year" binding:"required"`
}

// CalculateHolidaysRequest represents query parameters for calculating holidays by date and name
type CalculateHolidaysRequest struct {
	Date string `form:"date" binding:"required"`
	Name string `form:"name" binding:"required"`
}
