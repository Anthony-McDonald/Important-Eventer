package main

import (
	"encoding/json"
	"net/http"
)

// handler handles HTTP requests and returns upcoming calendar events in JSON format.
func handler(w http.ResponseWriter, r *http.Request) {

	cal, err := getCalendar()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	limit := config.Calendar.EventsToReturn
	if limit <= 0 {
		limit = 5
	}

	icsEvents, err := findNextEvents(cal, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	events, err := getFormedEventsFromIcsEvents(icsEvents)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(Response{
		Events: events,
	})
}
