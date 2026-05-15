package main

import (
	"os"
	"testing"
)

// TestLoadConfig tests that our config loader correctly handles environmental variables.
func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name           string
		envVars        map[string]string
		expectedErr    bool
		expectedPort   int
		expectedEvents int
		expectedTTL    int
		expectedURL    string
	}{
		{
			name: "loads config with defaults",
			envVars: map[string]string{
				"CALENDAR_URL": "https://example.com/calendar.ics",
			},
			expectedErr:    false,
			expectedPort:   8080,
			expectedEvents: 4,
			expectedTTL:    1,
			expectedURL:    "https://example.com/calendar.ics",
		},
		{
			name: "loads overridden values",
			envVars: map[string]string{
				"SERVER_PORT":               "9090",
				"CALENDAR_URL":              "https://example.com/custom.ics",
				"CALENDAR_EVENTS_TO_RETURN": "10",
				"CALENDAR_REFRESH_MINUTES":  "15",
			},
			expectedErr:    false,
			expectedPort:   9090,
			expectedEvents: 10,
			expectedTTL:    15,
			expectedURL:    "https://example.com/custom.ics",
		},
		{
			name: "fails when calendar url missing",
			envVars: map[string]string{
				"SERVER_PORT": "8080",
			},
			expectedErr: true,
		},
		{
			name: "fails with invalid integer env",
			envVars: map[string]string{
				"CALENDAR_URL":             "https://example.com/calendar.ics",
				"CALENDAR_REFRESH_MINUTES": "not-a-number",
			},
			expectedErr: true,
		},
	}

	envKeys := []string{
		"SERVER_PORT",
		"CALENDAR_URL",
		"CALENDAR_EVENTS_TO_RETURN",
		"CALENDAR_REFRESH_MINUTES",
	}

	for _, tC := range tests {
		t.Run(tC.name, func(t *testing.T) {
			// Clear relevant env vars before each test
			for _, key := range envKeys {
				_ = os.Unsetenv(key)
			}

			// Set test env vars
			for key, value := range tC.envVars {
				_ = os.Setenv(key, value)
			}

			cfg, err := loadConfig()

			if tC.expectedErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if cfg.Server.Port != tC.expectedPort {
				t.Fatalf("expected port %d, got %d", tC.expectedPort, cfg.Server.Port)
			}

			if cfg.Calendar.CalendarURL != tC.expectedURL {
				t.Fatalf("expected URL %q, got %q", tC.expectedURL, cfg.Calendar.CalendarURL)
			}

			if cfg.Calendar.EventsToReturn != tC.expectedEvents {
				t.Fatalf(
					"expected EventsToReturn %d, got %d",
					tC.expectedEvents,
					cfg.Calendar.EventsToReturn,
				)
			}

			if cfg.Calendar.RefreshMinutes != tC.expectedTTL {
				t.Fatalf(
					"expected RefreshMinutes %d, got %d",
					tC.expectedTTL,
					cfg.Calendar.RefreshMinutes,
				)
			}
		})
	}
}
