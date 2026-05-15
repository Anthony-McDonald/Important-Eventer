package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	ics "github.com/arran4/golang-ical"
)

// TestgetCalendar ensures that we can parse ICS and form a calendar from it. It also tests the server fetching functionality of the getCalendar function.
func TestGetCalendar(t *testing.T) {
	validICS := `BEGIN:VCALENDAR
VERSION:2.0
BEGIN:VEVENT
UID:test-event
SUMMARY:Test Event
DTSTART:20260101T120000Z
DTEND:20260101T130000Z
END:VEVENT
END:VCALENDAR`

	tests := []struct {
		name             string
		responseBody     string
		refreshMinutes   int
		initialLastFetch time.Time
		initialCachedCal *ics.Calendar
		serverEnabled    bool
		expectError      bool
		expectedRequests int
		expectCacheReuse bool
	}{
		{
			name:             "fetches and caches calendar",
			responseBody:     validICS,
			refreshMinutes:   10,
			serverEnabled:    true,
			expectError:      false,
			expectedRequests: 1,
		},
		{
			name:             "uses cached calendar within ttl",
			responseBody:     validICS,
			refreshMinutes:   10,
			initialLastFetch: time.Now(),
			initialCachedCal: ics.NewCalendar(),
			serverEnabled:    true,
			expectError:      false,
			expectedRequests: 0,
			expectCacheReuse: true,
		},
		{
			name:             "refreshes expired cache",
			responseBody:     validICS,
			refreshMinutes:   1,
			initialLastFetch: time.Now().Add(-2 * time.Minute),
			initialCachedCal: ics.NewCalendar(),
			serverEnabled:    true,
			expectError:      false,
			expectedRequests: 1,
		},
		{
			name:             "returns parse error for invalid ics",
			responseBody:     "invalid-ics",
			refreshMinutes:   10,
			serverEnabled:    true,
			expectError:      true,
			expectedRequests: 1,
		},
		{
			name:             "returns http error",
			refreshMinutes:   10,
			serverEnabled:    false,
			expectError:      true,
			expectedRequests: 0,
		},
	}

	for _, tC := range tests {
		t.Run(tC.name, func(t *testing.T) {
			// Reset globals
			cachedCal = tC.initialCachedCal
			lastFetch = tC.initialLastFetch

			requestCount := 0

			var server *httptest.Server

			if tC.serverEnabled {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					requestCount++
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(tC.responseBody))
				}))
				defer server.Close()

				config.Calendar.CalendarURL = server.URL
			} else {
				config.Calendar.CalendarURL = "http://127.0.0.1:0"
			}

			config.Calendar.RefreshMinutes = tC.refreshMinutes

			originalCached := cachedCal

			cal, err := getCalendar()

			if tC.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				if cal != nil {
					t.Fatal("expected nil calendar on error")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if cal == nil {
					t.Fatal("expected calendar, got nil")
				}
			}

			if requestCount != tC.expectedRequests {
				t.Fatalf("expected %d requests, got %d", tC.expectedRequests, requestCount)
			}

			if tC.expectCacheReuse && cal != originalCached {
				t.Fatal("expected cached calendar instance to be reused")
			}
		})
	}
}
