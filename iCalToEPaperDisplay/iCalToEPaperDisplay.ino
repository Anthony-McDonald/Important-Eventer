#include <Arduino.h>
#include <WiFi.h>
#include <HTTPClient.h>
#include <ArduinoJson.h>
#include <time.h>

#include "epd_driver.h"
#include "firasans.h"
#include "esp_sleep.h"

// ---------------- WIFI ----------------
const char* ssid = "Vodafone90E0C0";
const char* password = "2yJbxsaPpZe9tRzn";

// ---------------- API ----------------
const char* endpoint = "http://192.168.1.169:7259/next";

// ---------------- DATA ----------------

struct Event {
    String title;
    String date;
    String time;
    int daysRemaining;
};

Event events[4];
int eventCount = 0;

// ---------------- DISPLAY ----------------
uint8_t *fb;

// ============================================================
// TEXT HELPERS (NEW)
// ============================================================

int getTextWidth(const String &text, const GFXfont *font) {
    int width = 0;

    for (uint16_t i = 0; i < text.length(); i++) {
        char c = text[i];

        // skip non-printable characters safely
        if (c < 32 || c > 126) continue;

        GFXglyph *glyph = &font->glyph[c - 32];
        width += glyph->advance_x;
    }

    return width;
}

String truncateWithEllipsis(const String &text, const GFXfont *font, int maxWidth) {

    if (getTextWidth(text, font) <= maxWidth) {
        return text;
    }

    String result = "";

    for (uint16_t i = 0; i < text.length(); i++) {

        char c = text[i];

        if (c < 32 || c > 126) continue;

        String test = result + c;

        if (getTextWidth(test + "...", font) > maxWidth) {
            return result + "...";
        }

        result += c;
    }

    return result + "...";
}

// ============================================================
// DISPLAY HELPERS
// ============================================================

void showMessage(const char* msg1, const char* msg2 = "") {

    epd_poweron();
    epd_clear();

    memset(fb, 0xFF, EPD_WIDTH * EPD_HEIGHT / 2);

    Rect_t area = epd_full_screen();

    int x = 20;
    int y = 100;

    write_string((GFXfont *)&FiraSans, (char *)msg1, &x, &y, fb);

    if (strlen(msg2) > 0) {
        x = 20;
        y = 170;
        write_string((GFXfont *)&FiraSans, (char *)msg2, &x, &y, fb);
    }

    epd_draw_grayscale_image(area, fb);
    epd_poweroff();
}

// ============================================================
// MAIN DISPLAY
// ============================================================

void drawScreen() {

    epd_poweron();
    epd_clear();

    memset(fb, 0xFF, EPD_WIDTH * EPD_HEIGHT / 2);
    Rect_t area = epd_full_screen();

    const int midX = EPD_WIDTH / 2;
    const int leftX = 20;
    const int dividerX = midX + 20;

    const int rightX = dividerX + 20;
    const int upX = rightX;
    const int lineSpacing = 5;

    int x, y;

    // ========================================================
    // LEFT SIDE - MAIN EVENT
    // ========================================================

    String t = events[0].title;
    const int maxLineLen = 24;

    String titleLine1 = t;
    String titleLine2 = "";

    if (t.length() > maxLineLen) {
        int breakPos = t.lastIndexOf(' ', maxLineLen);
        if (breakPos == -1) breakPos = maxLineLen;

        titleLine1 = t.substring(0, breakPos);
        titleLine2 = t.substring(breakPos + 1);
    }

    x = leftX;
    y = 50;

    write_string((GFXfont *)&FiraSans, (char *)titleLine1.c_str(), &x, &y, fb);

    if (titleLine2.length() > 0) {
        y += lineSpacing;
        x = leftX;
        write_string((GFXfont *)&FiraSans, (char *)titleLine2.c_str(), &x, &y, fb);
    }

    // ========================================================
    // DIVIDER
    // ========================================================

    int dividerY = y - 30;
    epd_draw_line(0, dividerY, dividerX, dividerY, 0, fb);
    epd_draw_line(dividerX, 0, dividerX, EPD_HEIGHT, 0, fb);

    // ========================================================
    // LEFT DETAILS
    // ========================================================

    y = dividerY + 50;
    x = leftX;

    String daysLine = "DAYS LEFT: " + String(events[0].daysRemaining);
    write_string((GFXfont *)&FiraSans, (char *)daysLine.c_str(), &x, &y, fb);

    y += 25;

    x = leftX;

    String dateLine = "Date: " + events[0].date;
    write_string((GFXfont *)&FiraSans, (char *)dateLine.c_str(), &x, &y, fb);

    y += 25;

    x = leftX;

    String timeLine = "Time: " + events[0].time;
    write_string((GFXfont *)&FiraSans, (char *)timeLine.c_str(), &x, &y, fb);

    // ========================================================
    // RIGHT SIDE
    // ========================================================

    epd_draw_line(dividerX, 70, EPD_WIDTH, 70, 0, fb);

    x = upX;
    y = 50;

    write_string((GFXfont *)&FiraSans, "UPCOMING", &x, &y, fb);

    const int rightMargin = 10;
    int maxWidth = EPD_WIDTH - upX - rightMargin;

    for (int i = 1; i < 4; i++) {
        if (i >= eventCount) break;

        y += 30;
        x = upX;

        String safeTitle = truncateWithEllipsis(events[i].title, &FiraSans, maxWidth);

        write_string((GFXfont *)&FiraSans,
                    (char *)safeTitle.c_str(),
                    &x, &y, fb);

        y += 15;
        x = upX;

        String d = events[i].date;
        int tpos2 = d.indexOf("T");
        if (tpos2 > 0) d = d.substring(0, tpos2);

        String dateLabel = "Date: " + d;

        write_string((GFXfont *)&FiraSans,
                    (char *)dateLabel.c_str(),
                    &x, &y, fb);

        // ======================================================
        // SEPARATOR LINE
        // ======================================================

        if (i < eventCount - 1 && i < 3) {
            int lineY = y - 22;
            epd_draw_line(dividerX, lineY, EPD_WIDTH - 10, lineY, 0, fb);
        }
    }

    // ========================================================
    // RENDER
    // ========================================================

    epd_draw_grayscale_image(area, fb);
    epd_poweroff();
}

// ============================================================
// WIFI + FETCH
// ============================================================

bool fetchData() {

    WiFi.begin(ssid, password);

    unsigned long start = millis();
    while (WiFi.status() != WL_CONNECTED) {
        delay(300);
        if (millis() - start > 15000) return false;
    }

    HTTPClient http;
    http.begin(endpoint);

    int code = http.GET();
    if (code <= 0) return false;

    String payload = http.getString();

    StaticJsonDocument<1024> doc;
    DeserializationError err = deserializeJson(doc, payload);
    if (err) return false;

    JsonArray arr = doc["events"];

    eventCount = 0;

    for (JsonObject obj : arr) {

        if (eventCount >= 4) break;

        events[eventCount].title = String((const char*)obj["title"]);
        events[eventCount].date = String((const char*)obj["date"]);
        events[eventCount].time = String((const char*)obj["time"]);
        events[eventCount].daysRemaining = obj["days_remaining"];

        eventCount++;
    }

    http.end();
    return true;
}

// ============================================================
// SLEEP LOGIC (8AM WAKE)
// ============================================================

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

    epd_poweroff();

    esp_sleep_enable_timer_wakeup((uint64_t)secondsUntil8 * 1000000ULL);
    esp_deep_sleep_start();
}

// ============================================================
// SETUP
// ============================================================

void setup() {

    Serial.begin(115200);

    epd_init();

    fb = (uint8_t *)ps_calloc(1, EPD_WIDTH * EPD_HEIGHT / 2);

    if (!fb) {
        Serial.println("Framebuffer fail");
        while (1);
    }

    if (fetchData() && eventCount > 0) {
        drawScreen();
    } else {
        showMessage("WiFi Failed", "Will retry tomorrow");
    }

    configTime(0, 0, "pool.ntp.org");

    struct tm timeinfo;
    getLocalTime(&timeinfo);

    goToSleepUntil8AM();
}

void loop() {}