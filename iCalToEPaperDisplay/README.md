# LilyGo T5 Event Display with WiFi and HTTP

This project runs on the LilyGo T5 ESP32 e-paper development board and displays upcoming event information fetched from a remote HTTP API.

The system is designed for low-power, always-on event display use cases such as calendars, reminders, and dashboards.

---

## Features

- WiFi connectivity (ESP32-based)
- HTTP GET request to fetch event data
- JSON parsing of event payloads
- Native e-paper rendering (no LCD required)
- Low power consumption (ideal for battery/USB-powered displays)
- Simple event display layout optimized for e-paper

---

## Hardware Requirements

- LilyGo T5 e-paper ESP32 board
- USB-C cable (for flashing + power)
- WiFi network access

---

## Wiring

No external wiring is required.

The LilyGo T5 includes:

- ESP32 microcontroller
- Built-in e-paper display
- Power management circuitry
- USB interface

Everything is integrated on the board.

---

## Setup

### 1. Configure WiFi

```
const char* ssid = "<SSID>";
const char* password = "<PASSWORD>";
```

---

### 2. Configure API Endpoint

```
const char* endpoint = "http://<DNS>:<PORT>/next";
```

This endpoint should return a JSON payload containing event data.

See [here](../iCalToJson/README.md) for my implementation.

---

### 3. Install Required Libraries

Install the following Arduino libraries:

- WiFi (ESP32 core)
- HTTPClient
- ArduinoJson
- GxEPD2 (e-paper driver)
- Adafruit GFX

### 4. Flashing the Board

These steps assume that you have already installed the relevant libraries, and chosen the relevant flash settings. See [here](https://github.com/Xinyuan-LilyGO/LilyGo-EPD47?) for information on how to do that.

1. Connect the board via the USB cable
2. Select ESP32S3 Dev Module as your board in Arduino IDE or PlatformIO
3. Set correct COM port
4. Press and hold the BOOT(IO0) button , While still pressing the BOOT(IO0) button, press RST
5. Release the RST
6. Release the BOOT(IO0) button
7. Upload sketch

---

### Refresh Logic

On startup and subsequently at 8am daily, the code fetches a new event list and displays it.

---

## Potential Improvements

- Add OTA updates for firmware
- Add WiFi reconnection backoff