# Arduino LCD Display with WiFi and HTTP

This Arduino project displays event information on an LCD screen, fetching data from a remote HTTP endpoint. It includes scrolling title display, and toggles between days remaining and event date on the second row of the LCD.

## Features

- Connects to WiFi automatically
- Fetches event data via HTTP GET request
- Displays scrolling title on the first row
- Toggles between days remaining and event date on the second row
- Handles JSON parsing and error states
- Automatically refreshes data every 24 minutes

## Hardware Requirements

- Arduino board (tested with ESP32)
- 16x2 LCD display (with 4-bit interface)
- Jumper wires
- WiFi network access

## Wiring

The LCD is connected using the following pins:

| LCD Pin | Arduino Pin |
|---------|-------------|
| RS      | 13          |
| Enable  | 14          |
| D4      | 27          |
| D5      | 26          |
| D6      | 25          |
| D7      | 33          |

## Setup

1. Update WiFi credentials:
```cpp
const char* ssid = "<SSID>";
const char* password = "<PASSWORD>";
```
Update endpoint URL:
```cpp
const char* endpoint = "http://<DNS>:<PORT>/next";
```

Upload the code to your Arduino board.

## Code Structure

### Main Components

- WiFi Connection: Automatically connects to the specified network
- HTTP Client: Fetches data from the configured endpoint
- JSON Parser: Parses incoming JSON data
- LCD Display: Shows scrolling title and toggling information
- Timing: Handles refresh intervals and display updates

### HTTP Response Format

Expected JSON format:

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

### Display Behavior

- Row 1: Shows the event title with scrolling animation for long titles
- Row 2: Alternates between:
    - "Days: X" (where X is days remaining)
    - "Date: YYYY-MM-DD" (event date)