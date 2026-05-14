package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	ics "github.com/arran4/golang-ical"
)

// cachedCal is the in-memory cached calendar instance.
var cachedCal *ics.Calendar

// lastFetch tracks when the calendar was last fetched.
var lastFetch time.Time

// cacheMu protects cached calendar state for concurrent access.
var cacheMu sync.Mutex

// getCalendar fetches and caches the remote ICS calendar with TTL-based refresh.
func getCalendar() (*ics.Calendar, error) {
	cacheMu.Lock()
	defer cacheMu.Unlock()

	if cachedCal != nil &&
		time.Since(lastFetch) < time.Duration(config.Calendar.RefreshMinutes)*time.Minute {
		return cachedCal, nil
	}

	resp, err := http.Get(config.Calendar.URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	cal, err := ics.ParseCalendar(resp.Body)
	if err != nil {
		return nil, err
	}

	cachedCal = cal
	lastFetch = time.Now()

	fmt.Printf("events loaded: %d\n", len(cal.Events()))

	return cal, nil
}
