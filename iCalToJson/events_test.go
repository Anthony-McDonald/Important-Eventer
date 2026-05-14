package main

import (
	"testing"
	"time"

	ics "github.com/arran4/golang-ical"
)

// TestFindNextEvents tests the event filtering logic.
func TestFindNextEvents(t *testing.T) {
	now := time.Now()

	createEvent := func(summary string, start, end time.Time) *ics.VEvent {
		event := ics.NewEvent(summary)
		event.SetSummary(summary)
		event.SetStartAt(start)
		event.SetEndAt(end)
		return event
	}

	tests := []struct {
		name           string
		events         []*ics.VEvent
		limit          int
		expectedCount  int
		expectedTitles []string
		expectError    bool
	}{
		{
			name: "returns sorted upcoming events",
			events: []*ics.VEvent{
				createEvent(
					"later",
					now.Add(48*time.Hour),
					now.Add(49*time.Hour),
				),
				createEvent(
					"soon",
					now.Add(24*time.Hour),
					now.Add(25*time.Hour),
				),
			},
			limit:          10,
			expectedCount:  2,
			expectedTitles: []string{"soon", "later"},
			expectError:    false,
		},
		{
			name: "applies limit",
			events: []*ics.VEvent{
				createEvent(
					"one",
					now.Add(24*time.Hour),
					now.Add(25*time.Hour),
				),
				createEvent(
					"two",
					now.Add(48*time.Hour),
					now.Add(49*time.Hour),
				),
			},
			limit:          1,
			expectedCount:  1,
			expectedTitles: []string{"one"},
			expectError:    false,
		},
		{
			name: "skips past events",
			events: []*ics.VEvent{
				createEvent(
					"past",
					now.Add(-24*time.Hour),
					now.Add(-23*time.Hour),
				),
				createEvent(
					"future",
					now.Add(24*time.Hour),
					now.Add(25*time.Hour),
				),
			},
			limit:          10,
			expectedCount:  1,
			expectedTitles: []string{"future"},
			expectError:    false,
		},
		{
			name: "returns error when no upcoming events",
			events: []*ics.VEvent{
				createEvent(
					"past",
					now.Add(-24*time.Hour),
					now.Add(-23*time.Hour),
				),
			},
			limit:         10,
			expectedCount: 0,
			expectError:   true,
		},
		{
			name:          "returns error with empty calendar",
			events:        []*ics.VEvent{},
			limit:         10,
			expectedCount: 0,
			expectError:   true,
		},
	}

	for _, tC := range tests {
		t.Run(tC.name, func(t *testing.T) {
			cal := ics.NewCalendar()

			for _, event := range tC.events {
				cal.AddVEvent(event)
			}

			result, err := findNextEvents(cal, tC.limit)

			if tC.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result) != tC.expectedCount {
				t.Fatalf(
					"expected %d events, got %d",
					tC.expectedCount,
					len(result),
				)
			}

			for i, expectedTitle := range tC.expectedTitles {
				prop := result[i].GetProperty(ics.ComponentPropertySummary)

				if prop == nil {
					t.Fatalf("event %d missing summary", i)
				}

				if prop.Value != expectedTitle {
					t.Fatalf(
						"expected title %q, got %q",
						expectedTitle,
						prop.Value,
					)
				}
			}
		})
	}
}

// TestGetFormedEventsFromIcsEvents tests if we can form ics events into our own representation which is returned to clients.
func TestGetFormedEventsFromIcsEvents(t *testing.T) {
	now := time.Now()

	createEvent := func(summary string, start, end time.Time) *ics.VEvent {
		event := ics.NewEvent(summary)
		event.SetSummary(summary)
		event.SetStartAt(start)
		event.SetEndAt(end)
		return event
	}

	tests := []struct {
		name          string
		events        []*ics.VEvent
		expectedCount int
		expectedTitle string
		expectError   bool
	}{
		{
			name: "converts events successfully",
			events: []*ics.VEvent{
				createEvent(
					"Test Event",
					now.Add(24*time.Hour),
					now.Add(26*time.Hour),
				),
			},
			expectedCount: 1,
			expectedTitle: "Test Event",
			expectError:   false,
		},
		{
			name:          "returns error for empty input",
			events:        []*ics.VEvent{},
			expectedCount: 0,
			expectError:   true,
		},
		{
			name: "handles multiple events",
			events: []*ics.VEvent{
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
			},
			expectedCount: 2,
			expectedTitle: "One",
			expectError:   false,
		},
	}

	for _, tC := range tests {
		t.Run(tC.name, func(t *testing.T) {
			result, err := getFormedEventsFromIcsEvents(tC.events)

			if tC.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result) != tC.expectedCount {
				t.Fatalf(
					"expected %d events, got %d",
					tC.expectedCount,
					len(result),
				)
			}

			if tC.expectedCount > 0 {
				if result[0].Title != tC.expectedTitle {
					t.Fatalf(
						"expected title %q, got %q",
						tC.expectedTitle,
						result[0].Title,
					)
				}

				if result[0].Date == "" {
					t.Fatal("expected formatted date")
				}

				if result[0].Time == "" {
					t.Fatal("expected formatted time")
				}
			}
		})
	}
}
