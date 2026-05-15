package main

import (
	"embed"
	"encoding/json"
	"net/http"
)

// MakeNextHandler forms the handler that handles HTTP requests and returns upcoming calendar events in JSON format.
func MakeNextHandler(embeddingVectorURL string, baseHostUrl string, iconSet embed.FS) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
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

		events, err := getFormedEventsFromIcsEvents(embeddingVectorURL, baseHostUrl, iconSet, icsEvents)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(Response{
			Events: events,
		})
	}
}
