# Calendar Event Tracker

A Go application that fetches and tracks the next upcoming event from an iCalendar (ICS) URL, returning the event title, days remaining, and date in JSON format.

## Features

- Fetches calendar data from a remote ICS URL
- Caches calendar data to avoid frequent HTTP requests
- Finds the next upcoming event after the current time
- Returns event details in JSON format via HTTP endpoint
- Environment-based configuration

## Installation

### Prerequisites

- Go 1.21 or higher

## Configuration

The application uses environment variables for configuration:


| Environment Variable            | Description                                  |       Default Value        |
|-------------------------------|------------------------------------------------|----------------------------|
| `SERVER_URL`                  | URL used to inform image hosting               |    `http://127.0.0.1"`     |
| `SERVER_PORT`                 | Port to run the HTTP server on                 |          `8080`            |
| `CALENDAR_URL`                | URL to the iCalendar (ICS) file                |        (required)          |
| `CALENDAR_EVENTS_TO_RETURN`   | How many of the fetched events to return       |           `4`              |
| `CALENDAR_REFRESH_MINUTES`    | How often to refresh the calendar data         |           `1`              |
| `EMBEDDING_VECTOR_URL`        | URL to request to get an embedding vector      |        (required)          |

## Usage

Start the server:

```bash
export CALENDAR_URL="https://example.com/calendar.ics"
# All other environmental variables are optional.
./calendar-event-tracker
```

### Endpoints

The application exposes one endpoint:

- `GET /next` - Returns the next upcoming event in JSON format

### Response Example

```json
{
  "events": [
    {
      "title": "Event title",
      "days_remaining": 0,
      "date": "MM-DD-YYYY",
      "time": "HH:MM-HH:MM"
    }
  ]
}
```

## Example Calendar URL

You can use any calendar URLs that support ICS such as Google Calendar, Outlook Calendar or personally I went the locally hosted route with Radicale :).

- Google Calendar ICS export
- Outlook Calendar ICS export
- 

Example:

```bash
export CALENDAR_URL="https://calendar.google.com/calendar/ical/your-calendar-id/basic.ics"
```

## Cache Behavior

The application caches the calendar data for a configurable duration (`CALENDAR_REFRESH_MINUTES`). If the cached data is still valid, it will not fetch new data from the URL.

## Error Handling

- If the calendar URL is not set, the application will panic.
- If the calendar data cannot be fetched or parsed, a 500 error is returned.
- If no upcoming events are found, a 404 error is returned.