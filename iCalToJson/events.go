package main

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"time"

	ics "github.com/arran4/golang-ical"
)

// findNextEvents returns the next upcoming events sorted by start time up to a limit.
func findNextEvents(cal *ics.Calendar, limit int) ([]*ics.VEvent, error) {

	now := time.Now()

	type timedEvent struct {
		event *ics.VEvent
		start time.Time
	}

	var events []timedEvent

	skippedPast := 0
	skippedError := 0

	for _, event := range cal.Events() {

		start, err := event.GetStartAt()
		if err != nil {
			skippedError++
			continue
		}

		if start.Before(now) {
			skippedPast++
			continue
		}

		events = append(events, timedEvent{
			event: event,
			start: start,
		})
	}

	fmt.Printf("skipped past=%d invalid=%d\n", skippedPast, skippedError)

	if len(events) == 0 {
		return nil, fmt.Errorf("no upcoming events")
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].start.Before(events[j].start)
	})

	if limit > len(events) {
		limit = len(events)
	}

	result := make([]*ics.VEvent, 0, limit)
	for i := 0; i < limit; i++ {
		result = append(result, events[i].event)
	}

	return result, nil
}

// getFormedEventsFromIcsEvents converts ICS events into API response-ready Event structs.
func getFormedEventsFromIcsEvents(embeddingVectorURL string, baseHostUrl string, iconSet embed.FS, events []*ics.VEvent) ([]Event, error) {

	if len(events) == 0 {
		return nil, errors.New("no events provided")
	}

	result := make([]Event, 0, len(events))
	var errs []error

	for _, event := range events {

		start, err := event.GetStartAt()
		if err != nil {
			errs = append(errs, err)
			continue
		}

		end, err := event.GetEndAt()
		if err != nil {
			errs = append(errs, err)
			continue
		}

		title := ""
		if prop := event.GetProperty(ics.ComponentPropertySummary); prop != nil {
			title = prop.Value
		}

		days := int(time.Until(start).Hours()/24) + 1

		time := fmt.Sprintf("%s-%s", start.Format("15:04"), end.Format("15:04"))

		imageToUse := chooseImage(embeddingVectorURL, title)

		path, err := getIconPath(iconSet, imageToUse)

		if err != nil {
			return []Event{}, fmt.Errorf("form events from ics events: %w", err)
		}

		fullPath := fmt.Sprintf("%s%s", baseHostUrl, path)

		if err != nil {
			return []Event{}, fmt.Errorf("find image %q:%w", imageToUse, err)
		}

		result = append(result, Event{
			Title:         title,
			RelevantImage: fullPath,
			DaysRemaining: days,
			Date:          start.Format("02-01-2006"),
			Time:          time,
		})
	}

	if len(errs) > 0 {
		return result, errors.Join(errs...)
	}

	return result, nil
}

// getIconPath finds the file with name 'target' in fsys and returns its URL path.
func getIconPath(fsys embed.FS, target string) (string, error) {
	var foundPath string
	var found bool

	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if filepath.Base(path) == target {
			foundPath = path
			found = true
			return fs.SkipAll
		}

		return nil
	})

	if err != nil {
		return "", err
	}
	if !found {
		return "", fmt.Errorf("file %s not found", target)
	}

	return foundPath, nil
}
