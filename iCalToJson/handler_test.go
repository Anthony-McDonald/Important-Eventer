package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	ics "github.com/arran4/golang-ical"
)

// TestHandler tests if our sole handler (that handles returning event information) can fetch and return data as expected.
func TestHandler(t *testing.T) {
	now := time.Now()

	createCalendar := func(events []*ics.VEvent) *ics.Calendar {
		cal := ics.NewCalendar()

		for _, event := range events {
			cal.AddVEvent(event)
		}

		return cal
	}

	createEvent := func(summary string, start, end time.Time) *ics.VEvent {
		event := ics.NewEvent(summary)
		event.SetSummary(summary)
		event.SetStartAt(start)
		event.SetEndAt(end)
		return event
	}

	tests := []struct {
		name               string
		calendar           *ics.Calendar
		eventsToReturn     int
		expectedStatus     int
		expectedEventCount int
		expectContentType  bool
	}{
		{
			name: "returns upcoming events",
			calendar: createCalendar([]*ics.VEvent{
				createEvent(
					"Test Event",
					now.Add(24*time.Hour),
					now.Add(25*time.Hour),
				),
			}),
			eventsToReturn:     5,
			expectedStatus:     http.StatusOK,
			expectedEventCount: 1,
			expectContentType:  true,
		},
		{
			name: "returns 404 when no upcoming events",
			calendar: createCalendar([]*ics.VEvent{
				createEvent(
					"Past Event",
					now.Add(-24*time.Hour),
					now.Add(-23*time.Hour),
				),
			}),
			eventsToReturn:     5,
			expectedStatus:     http.StatusNotFound,
			expectedEventCount: 0,
			expectContentType:  false,
		},
		{
			name: "uses default limit when config limit invalid",
			calendar: createCalendar([]*ics.VEvent{
				createEvent(
					"One",
					now.Add(24*time.Hour),
					now.Add(25*time.Hour),
				),
				createEvent(
					"Two",
					now.Add(48*time.Hour),
					now.Add(49*time.Hour),
				),
			}),
			eventsToReturn:     0,
			expectedStatus:     http.StatusOK,
			expectedEventCount: 2,
			expectContentType:  true,
		},
	}

	for _, tC := range tests {
		t.Run(tC.name, func(t *testing.T) {
			// Mock cache state
			cachedCal = tC.calendar
			lastFetch = time.Now()

			config.Calendar.EventsToReturn = tC.eventsToReturn
			config.Calendar.RefreshMinutes = 10

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()

			handler(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			if res.StatusCode != tC.expectedStatus {
				t.Fatalf(
					"expected status %d, got %d",
					tC.expectedStatus,
					res.StatusCode,
				)
			}

			if tC.expectContentType {
				contentType := res.Header.Get("Content-Type")

				if contentType != "application/json" {
					t.Fatalf(
						"expected Content-Type application/json, got %q",
						contentType,
					)
				}

				var response Response

				if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if len(response.Events) != tC.expectedEventCount {
					t.Fatalf(
						"expected %d events, got %d",
						tC.expectedEventCount,
						len(response.Events),
					)
				}
			}
		})
	}
}
