package model

// CalculateRequest represents query parameters for the calculator endpoint.
type CalculateRequest struct {
	Date string `form:"date" binding:"required"`
	Name string `form:"name" binding:"required"`
}

// CalculateResponse represents the result of the calculation.
type CalculateResponse struct {
	HolidayName string `json:"holiday_name"`
	HolidayDate string `json:"holiday_date"`
	FromDate    string `json:"from_date"`
	DaysLeft    int    `json:"days_left"`
	Source      string `json:"source"`
}
