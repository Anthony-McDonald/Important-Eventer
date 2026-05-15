package main

// Event represents a simplified calendar event returned in API responses.
type Event struct {
	Title         string `json:"title"`
	RelevantImage string `json:"image"`
	DaysRemaining int    `json:"days_remaining"`
	Date          string `json:"date"`
	Time          string `json:"time"`
}

// Response defines the JSON payload returned by the API.
type Response struct {
	Events []Event `json:"events"`
}
