#include <Arduino.h>
#include <WiFi.h>
#include <HTTPClient.h>
#include <ArduinoJson.h>
#include <time.h>

#include "epd_driver.h"
#include "firasans.h"
#include "esp_sleep.h"
#include <JPEGDEC.h>


#define IMG_W 256
#define IMG_H 256

#define DRAW_W 256
#define DRAW_H 256

JPEGDEC jpeg;

// Separate buffers: one for the raw JPEG download, one for decoded grayscale pixels.
uint8_t *jpegDownloadBuf = nullptr;
uint8_t  decodedImg[IMG_W * IMG_H];

// ---- WIFI ----
const char* ssid     = "Vodafone90E0C0";
const char* password = "2yJbxsaPpZe9tRzn";

// ---- API ----
const char* endpoint = "http://192.168.1.169:7259/next";

// ---- DISPLAY ----
uint8_t *fb;

// ---- DATA ----

// Event represents a calendar event returned by the API.
struct Event {
    String title;
    String image;
    String date;
    String time;
    int    daysRemaining;
};

Event events[4];
int   eventCount = 0;

int  drawJPEG(JPEGDRAW *pDraw);
void drawImageBottomLeft();

// =============================================================================
// TEXT UTILITIES
// =============================================================================

// getTextWidth calculates the rendered pixel width of a string using the given font.
int getTextWidth(const String &text, const GFXfont *font) {
    int width = 0;
    for (uint16_t i = 0; i < text.length(); i++) {
        char c = text[i];
        if (c < 32 || c > 126) continue;
        width += font->glyph[c - 32].advance_x;
    }
    return width;
}

// truncateWithEllipsis shortens text to fit within maxWidth pixels and appends "...".
String truncateWithEllipsis(const String &text, const GFXfont *font, int maxWidth) {
    if (getTextWidth(text, font) <= maxWidth) return text;

    String result = "";
    for (uint16_t i = 0; i < text.length(); i++) {
        char c = text[i];
        if (c < 32 || c > 126) continue;
        if (getTextWidth(result + c + "...", font) > maxWidth) {
            return result + "...";
        }
        result += c;
    }
    return result + "...";
}

// =============================================================================
// DISPLAY
// =============================================================================

// showMessage displays one or two lines of text centred vertically on screen.
void showMessage(const char *msg1, const char *msg2 = "") {
    epd_poweron();
    epd_clear();
    memset(fb, 0xFF, EPD_WIDTH * EPD_HEIGHT / 2);

    int x = 20, y = 100;
    write_string((GFXfont *)&FiraSans, (char *)msg1, &x, &y, fb);

    if (strlen(msg2) > 0) {
        x = 20; y = 170;
        write_string((GFXfont *)&FiraSans, (char *)msg2, &x, &y, fb);
    }

    epd_draw_grayscale_image(epd_full_screen(), fb);
    epd_poweroff();
}

// drawScreen renders the main dashboard: primary event on the left, upcoming
// events on the right, and the event icon in the bottom-left corner.
void drawScreen() {
    epd_poweron();
    epd_clear();
    memset(fb, 0xFF, EPD_WIDTH * EPD_HEIGHT / 2);

    const int leftX    = 20;
    const int dividerX = EPD_WIDTH / 2 + 20;
    const int rightX   = dividerX + 20;

    int x, y;

    // Primary event title (word-wrapped to two lines).
    String t = events[0].title;
    const int maxLineLen = 24;
    String titleLine1 = t;
    String titleLine2 = "";

    if ((int)t.length() > maxLineLen) {
        int breakPos = t.lastIndexOf(' ', maxLineLen);
        if (breakPos == -1) breakPos = maxLineLen;
        titleLine1 = t.substring(0, breakPos);
        titleLine2 = t.substring(breakPos + 1);
    }

    x = leftX; y = 50;
    write_string((GFXfont *)&FiraSans, (char *)titleLine1.c_str(), &x, &y, fb);

    if (titleLine2.length() > 0) {
        y += 5; x = leftX;
        write_string((GFXfont *)&FiraSans, (char *)titleLine2.c_str(), &x, &y, fb);
    }

    // Horizontal rule beneath title / vertical divider.
    int dividerY = y - 30;
    epd_draw_line(0,        dividerY, dividerX,    dividerY,   0, fb);
    epd_draw_line(dividerX, 0,        dividerX,    EPD_HEIGHT, 0, fb);

    // Primary event details.
    y = dividerY + 50; x = leftX;
    String daysLine;
    if (events[0].daysRemaining == 0) {
        daysLine = "TODAY!";
    } else {
        daysLine = "DAYS LEFT: " + String(events[0].daysRemaining);
    }
    write_string((GFXfont *)&FiraSans, (char *)daysLine.c_str(), &x, &y, fb);

    y += 25; x = leftX;
    String dateLine = "Date: " + events[0].date;
    write_string((GFXfont *)&FiraSans, (char *)dateLine.c_str(), &x, &y, fb);

    y += 25; x = leftX;
    String timeLine = "Time: " + events[0].time;
    write_string((GFXfont *)&FiraSans, (char *)timeLine.c_str(), &x, &y, fb);

    // Upcoming events panel.
    epd_draw_line(dividerX, 70, EPD_WIDTH, 70, 0, fb);

    x = rightX; y = 50;
    write_string((GFXfont *)&FiraSans, "UPCOMING", &x, &y, fb);

    const int maxWidth = EPD_WIDTH - rightX - 10;

    for (int i = 1; i < 4; i++) {
        if (i >= eventCount) break;

        y += 30; x = rightX;
        String safeTitle = truncateWithEllipsis(events[i].title, &FiraSans, maxWidth);
        write_string((GFXfont *)&FiraSans, (char *)safeTitle.c_str(), &x, &y, fb);

        y += 15; x = rightX;
        String d = events[i].date;
        int tpos = d.indexOf('T');
        if (tpos > 0) d = d.substring(0, tpos);
        String dateLabel = "Date: " + d;
        write_string((GFXfont *)&FiraSans, (char *)dateLabel.c_str(), &x, &y, fb);

        if (i < eventCount - 1 && i < 3) {
            epd_draw_line(dividerX, y - 22, EPD_WIDTH - 10, y - 22, 0, fb);
        }
    }

    // Event icon.
    drawImageBottomLeft();

    epd_draw_grayscale_image(epd_full_screen(), fb);
    epd_poweroff();
}

// drawImageBottomLeft draws the fetched image at the bottom left of the screen.
void drawImageBottomLeft() {
    int startX = 50;
    int startY = EPD_HEIGHT - DRAW_H + 90;

    for (int row = 0; row < DRAW_H; row++) {
        for (int col = 0; col < DRAW_W; col++) {

            uint8_t gray  = decodedImg[row * IMG_W + col];
            uint8_t pixel = gray;

            epd_draw_pixel(startX + col, startY + row, pixel, fb);
        }
    }
}

// =============================================================================
// IMAGE FETCHING & JPEG DECODING
// =============================================================================

// drawJPEG is the JPEGDEC pixel callback. It converts each RGB565 block to
// 8-bit grayscale and stores it in decodedImg for later blitting.
int drawJPEG(JPEGDRAW *pDraw) {
    uint16_t *pixels = (uint16_t *)pDraw->pPixels;

    for (int row = 0; row < pDraw->iHeight; row++) {
        for (int col = 0; col < pDraw->iWidth; col++) {
            uint16_t c = pixels[row * pDraw->iWidth + col];

            // Extract RGB565 channels.
            uint8_t r5 = (c >> 11) & 0x1F;
            uint8_t g6 = (c >>  5) & 0x3F;
            uint8_t b5 =  c        & 0x1F;

            // Scale each channel to 8 bits.
            uint8_t r8 = (r5 << 3) | (r5 >> 2);
            uint8_t g8 = (g6 << 2) | (g6 >> 4);
            uint8_t b8 = (b5 << 3) | (b5 >> 2);

            // Rec. 601 luma.
            uint8_t gray = (uint8_t)((r8 * 77u + g8 * 150u + b8 * 29u) >> 8);

            int px = pDraw->x + col;
            int py = pDraw->y + row;

            if (px >= 0 && px < IMG_W && py >= 0 && py < IMG_H) {
                decodedImg[py * IMG_W + px] = gray;
            }
        }
    }
    return 1;
}

// fetchImage downloads a JPEG from url, decodes it into decodedImg, and
// returns true on success.
bool fetchImage(const String &url) {
    HTTPClient http;
    http.begin(url);

    int code = http.GET();
    if (code != 200) {
        Serial.printf("fetchImage: HTTP %d\n", code);
        http.end();
        return false;
    }

    int len = http.getSize();
    if (len <= 0) {
        Serial.println("fetchImage: invalid content-length");
        http.end();
        return false;
    }

    // Allocate or reallocate the download buffer in PSRAM.
    if (jpegDownloadBuf) {
        free(jpegDownloadBuf);
        jpegDownloadBuf = nullptr;
    }
    jpegDownloadBuf = (uint8_t *)ps_malloc(len);
    if (!jpegDownloadBuf) {
        Serial.println("fetchImage: ps_malloc failed");
        http.end();
        return false;
    }

    // Read the full JPEG payload into the buffer.
    WiFiClient *stream = http.getStreamPtr();
    int received = 0;
    unsigned long deadline = millis() + 10000;
    while (http.connected() && received < len && millis() < deadline) {
        if (stream->available()) {
            jpegDownloadBuf[received++] = stream->read();
        } else {
            delay(1);
        }
    }
    http.end();

    if (received != len) {
        Serial.printf("fetchImage: expected %d bytes, got %d\n", len, received);
        return false;
    }
    Serial.printf("fetchImage: downloaded %d bytes\n", received);

    // Decode into decodedImg (default to white in case the image is smaller).
    memset(decodedImg, 0xFF, sizeof(decodedImg));

    uint8_t centre = decodedImg[(IMG_H / 2) * IMG_W + (IMG_W / 2)];
    Serial.printf("Centre pixel gray value: %d\n", centre);

    // Ensure pixel format matches what drawJPEG expects.
    jpeg.setPixelType(RGB565_LITTLE_ENDIAN);

    if (!jpeg.openRAM(jpegDownloadBuf, received, drawJPEG)) {
        Serial.println("fetchImage: JPEG open failed");
        return false;
    }

    jpeg.decode(0, 0, 0);
    jpeg.close();

    // Sanity check — if every pixel is white the decode callback never fired.
    int nonWhite = 0;
    for (int i = 0; i < IMG_W * IMG_H; i++) {
        if (decodedImg[i] != 0xFF) nonWhite++;
    }
    Serial.printf("fetchImage: decoded, non-white pixels = %d\n", nonWhite);

    return true;
}

// =============================================================================
// DATA FETCHING
// =============================================================================

// fetchData connects to WiFi, retrieves event JSON from the API, and
// downloads the icon for the primary event. Returns true on success.
bool fetchData() {
    WiFi.begin(ssid, password);

    unsigned long start = millis();
    while (WiFi.status() != WL_CONNECTED) {
        delay(300);
        if (millis() - start > 15000) {
            Serial.println("fetchData: WiFi timeout");
            return false;
        }
    }
    Serial.println("fetchData: WiFi connected");

    HTTPClient http;
    http.begin(endpoint);

    int code = http.GET();
    if (code <= 0) {
        Serial.printf("fetchData: HTTP error %d\n", code);
        http.end();
        return false;
    }

    String payload = http.getString();
    http.end();

    StaticJsonDocument<1024> doc;
    DeserializationError err = deserializeJson(doc, payload);
    if (err) {
        Serial.printf("fetchData: JSON parse error: %s\n", err.c_str());
        return false;
    }

    JsonArray arr = doc["events"];
    eventCount = 0;

    for (JsonObject obj : arr) {
        if (eventCount >= 4) break;
        events[eventCount].title         = String((const char *)obj["title"]);
        events[eventCount].image         = String((const char *)obj["image"]);
        events[eventCount].date          = String((const char *)obj["date"]);
        events[eventCount].time          = String((const char *)obj["time"]);
        events[eventCount].daysRemaining = obj["days_remaining"];
        eventCount++;
    }

    Serial.printf("fetchData: got %d events\n", eventCount);

    if (eventCount > 0 && events[0].image.length() > 0) {
        Serial.printf("fetchData: fetching image: %s\n", events[0].image.c_str());
        fetchImage(events[0].image);
    }

    return true;
}

// =============================================================================
// SLEEP
// =============================================================================

// goToSleepUntil8AM calculates the seconds until 08:00 local time and enters
// ESP32 deep sleep for that duration.
void goToSleepUntil8AM() {
    time_t now;
    struct tm timeinfo;
    time(&now);
    localtime_r(&now, &timeinfo);

    int h = timeinfo.tm_hour;
    int m = timeinfo.tm_min;

    int secondsUntil8;
    if (h < 8) {
        secondsUntil8 = (8 - h) * 3600 - m * 60;
    } else {
        secondsUntil8 = ((24 - h) + 8) * 3600 - m * 60;
    }

    Serial.printf("Sleeping for %d seconds\n", secondsUntil8);
    epd_poweroff();
    esp_sleep_enable_timer_wakeup((uint64_t)secondsUntil8 * 1000000ULL);
    esp_deep_sleep_start();
}

// =============================================================================
// ENTRY POINTS
// =============================================================================

void setup() {
    Serial.begin(115200);
    epd_init();

    fb = (uint8_t *)ps_calloc(1, EPD_WIDTH * EPD_HEIGHT / 2);
    if (!fb) {
        Serial.println("setup: framebuffer allocation failed");
        while (1);
    }

    if (fetchData() && eventCount > 0) {
        drawScreen();
    } else {
        showMessage("WiFi Failed", "Will retry tomorrow");
    }

    // Sync time after display update so the sleep calculation is accurate.
    configTime(0, 0, "pool.ntp.org");
    struct tm timeinfo;
    getLocalTime(&timeinfo);

    goToSleepUntil8AM();
}

// loop is unused — the device sleeps immediately after setup() completes.
void loop() {}