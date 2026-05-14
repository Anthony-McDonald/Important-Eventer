#include <WiFi.h>
#include <HTTPClient.h>
#include <ArduinoJson.h>
#include <LiquidCrystal.h>

// ---- WiFi credentials ----
const char* ssid = "<SSID>";
const char* password = "<PASSWORD>";

// ---- Endpoint ----
const char* endpoint = "http://<DNS>:<PORT>/next";

// ---- LCD (RS, E, D4, D5, D6, D7) ----
LiquidCrystal lcd(13, 14, 27, 26, 25, 33);

// ---- Scroll state (title) ----
String currentTitle = "";
int scrollIndex = 0;
unsigned long lastScrollTime = 0;
const int scrollDelay = 600;

// ---- Display state (row 2 toggle) ----
int daysRemaining = 0;
String eventDate = "";
bool showDays = true;
unsigned long lastToggleTime = 0;
const unsigned long toggleInterval = 2500;

// ---- Timing ----
unsigned long lastFetchTime = 0;
const unsigned long fetchInterval = 1440000;

void setup() {
  Serial.begin(115200);
  delay(1000);

  lcd.begin(16, 2);
  lcd.clear();
  lcd.print("Connecting WiFi");

  WiFi.begin(ssid, password);

  while (WiFi.status() != WL_CONNECTED) {
    delay(500);
    Serial.print(".");
  }

  Serial.println("\nWiFi connected!");
  Serial.println(WiFi.localIP());

  lcd.clear();
  lcd.print("WiFi Connected");

  delay(1000);

  makeRequest();
}

void loop() {
  updateScroll();
  updateSecondRow();

  if (millis() - lastFetchTime > fetchInterval) {
    lastFetchTime = millis();
    makeRequest();
  }
}

// ---------------- HTTP + JSON ----------------
void makeRequest() {
  if (WiFi.status() != WL_CONNECTED) {
    lcd.clear();
    lcd.print("WiFi lost");
    return;
  }

  HTTPClient http;

  lcd.clear();
  lcd.print("Fetching...");

  http.begin(endpoint);

  int httpCode = http.GET();

  if (httpCode > 0) {
    String payload = http.getString();

    StaticJsonDocument<512> doc;
    DeserializationError error = deserializeJson(doc, payload);

    if (error) {
      lcd.clear();
      lcd.print("JSON error");
      return;
    }

    const char* title = doc["title"];
    daysRemaining = doc["days_remaining"];
    const char* date = doc["date"];

    currentTitle = String(title);
    eventDate = String(date);

    scrollIndex = 0;
    showDays = true;

    Serial.println(currentTitle);
    Serial.println(daysRemaining);
    Serial.println(eventDate);

  } else {
    lcd.clear();
    lcd.print("HTTP error");
  }

  http.end();
}

// ---------------- TITLE SCROLL ----------------
void updateScroll() {
  if (currentTitle.length() == 0) return;

  if (currentTitle.length() <= 16) {
    lcd.setCursor(0, 0);
    lcd.print(currentTitle);
    return;
  }

  if (millis() - lastScrollTime < scrollDelay) return;
  lastScrollTime = millis();

  String displayText;

  if (scrollIndex + 16 <= currentTitle.length()) {
    displayText = currentTitle.substring(scrollIndex, scrollIndex + 16);
  } else {
    int part1 = currentTitle.length() - scrollIndex;
    displayText = currentTitle.substring(scrollIndex);
    displayText += " ";
    displayText += currentTitle.substring(0, 16 - part1 - 1);
  }

  lcd.setCursor(0, 0);
  lcd.print(displayText);

  scrollIndex++;
  if (scrollIndex >= currentTitle.length()) {
    scrollIndex = 0;
  }
}

// ---------------- SECOND ROW TOGGLE ----------------
void updateSecondRow() {
  if (millis() - lastToggleTime < toggleInterval) return;
  lastToggleTime = millis();

  lcd.setCursor(0, 1);
  lcd.print("                "); // clear row

  lcd.setCursor(0, 1);

  if (showDays) {
    lcd.print("Days: ");
    lcd.print(daysRemaining);
  } else {
    lcd.print("Date:");
    lcd.print(eventDate);
  }

  showDays = !showDays;
}