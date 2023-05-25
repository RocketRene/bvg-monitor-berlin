#include <WiFi.h>
#include <HTTPClient.h>
#include <ArduinoJson.h>
#include <TFT_eSPI.h>


const char* ssid = "Wokwi-GUEST";
const char* password = "";
const char* apiUrl = "https://bvg.fly.dev/";
const int buttonPin = 5;
bool isButtonPressed = false;

#define TFT_DC 2
#define TFT_CS 15
TFT_eSPI tft = TFT_eSPI(TFT_WIDTH, TFT_HEIGHT); // adjust width and height to match your display

void setup() {
  Serial.begin(115200);
  tft.begin();
  tft.setRotation(1); // adjust rotation to match your display
  tft.fillScreen(TFT_WHITE);
  pinMode(buttonPin, INPUT_PULLUP);
    // Connect to WiFi
  WiFi.begin(ssid, password);
  while (WiFi.status() != WL_CONNECTED) {
    delay(1000);
    Serial.println("Connecting to WiFi...");
    // Print to the TFT display
    tft.fillScreen(TFT_WHITE);
    tft.setCursor(0, 0);
    tft.setTextColor(TFT_RED);
    tft.setTextSize(5);
    tft.println("Connecting to WiFi...");
    
  }
}

void loop() {


  

  // Make a HTTP request to the API
  HTTPClient http;
  http.begin(apiUrl);
  int httpResponseCode = http.GET();
  if (httpResponseCode == HTTP_CODE_OK) {
    // Parse the JSON response
    String payload = http.getString();
    DynamicJsonDocument doc(1024);
    deserializeJson(doc, payload);
    JsonVariant bus = doc["bus"];
    JsonVariant weather = doc["wheather"];
    JsonVariant time = doc["timestamp"];
    Serial.println(payload);
    Serial.println();
    Serial.println("Weather data:");
    serializeJson(weather, Serial);
    Serial.println();
    Serial.println();
        
    // Display the data on the TFT display
    tft.fillScreen(TFT_WHITE);
    tft.setCursor(0, 0);
    tft.setTextColor(TFT_PURPLE);
    tft.setTextSize(1);
    tft.println();
    tft.setTextSize(4);
    //Current time
    tft.print(" Time:");
    tft.setTextColor(TFT_BLACK);
    tft.println(time.as<String>());
    tft.setTextSize(2);
    tft.setTextColor(TFT_BLACK);

    for (auto busStop : bus.as<JsonObject>()) {
      tft.print(" ");
      tft.print(busStop.key().c_str());
      tft.print(": ");
      for (auto time : busStop.value().as<JsonArray>()) {
        //if time is less then 1 minute, print in red
        if (time.as<int>() < 1) {
          tft.setTextColor(TFT_RED);
        }

        tft.print(time.as<int>());
        tft.print(" ");
        tft.setTextColor(TFT_BLACK);
      }
      tft.println();
      tft.println();

      
    }
   
    tft.setTextColor(TFT_BLUE);
    tft.setTextSize(4);
    tft.println(" Weather:");
    tft.setTextColor(TFT_BLACK);
    tft.setTextSize(2);
    tft.print(" Temperature: ");
    tft.print(weather["temperature"].as<float>());
    tft.print(" C");
    tft.println();
    tft.print(" Precipitation: ");
    tft.print(weather["precipitation_probability_max"].as<float>());
    tft.print(" %");
    tft.println();
    tft.print(" Sunset time: ");
    tft.print(weather["sunset"].as<String>());
    
    tft.println();
    tft.println();

    
        tft.setTextSize(1);
    tft.print(" Made for Olivia and Barack by Rene");
    tft.setTextColor(TFT_RED);
    tft.print("<3");

  } else {
    Serial.println("HTTP request failed.");
  }
  http.end();

  //while wating 5 seconds for the next request, check if the button is pressed
    // Check if button is pressed

  for (int i = 0; i < 50; i++) {    
    if (digitalRead(buttonPin) == LOW && !isButtonPressed) {
      isButtonPressed = true;
      Serial.println("Button pressed");
      HTTPClient http;
      http.begin("http://door.local");
      int httpResponseCode = http.GET();
      if (httpResponseCode == HTTP_CODE_OK) {
        Serial.println("Door opened");
        tft.fillScreen(ILI9341_BLACK);
        tft.setCursor(0, 0);
        tft.setTextColor(ILI9341_GREEN);
        tft.setTextSize(4);
        tft.println("Door opened");
      } else {
        Serial.println("Failed to open door");
        tft.fillScreen(ILI9341_BLACK);
        tft.setCursor(0, 0);
        tft.setTextColor(ILI9341_RED);
        tft.setTextSize(4);
        tft.println("Failed to open door");
      }
      http.end();
    } else if (digitalRead(buttonPin) == HIGH) {
      isButtonPressed = false;
    }
    delay(100);
  }
}
