package command

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestFormatWeatherResponse(t *testing.T) {
	// Test weather data
	weatherData := &OneCallResponse{
		Current: CurrentWeather{
			Temp:      20.5,
			FeelsLike: 22.3,
			WindSpeed: 3.2,
			Humidity:  65,
			Pressure:  1013,
			Clouds:    40,
			UVI:       5.2,
			Weather: []WeatherCondition{{
				Description: "partly cloudy",
			}},
		},
		Alerts: []Alert{{
			Event: "High Wind Warning",
		}},
	}

	result := formatWeatherResponse("Helsinki", "FI", weatherData)

	expectedParts := []string{
		"Helsinki, FI",
		"Temperature: 20.5°C",
		"feels like: 22.3°C",
		"wind: 3.2 m/s",
		"humidity: 65%",
		"pressure: 1013hPa",
		"cloudiness: 40%",
		"partly cloudy",
		"UV index: 5.2",
		"⚠️ High Wind Warning",
	}

	for _, part := range expectedParts {
		if !strings.Contains(result, part) {
			t.Errorf("Expected result to contain '%s', but got: %s", part, result)
		}
	}
}

func TestFormatWeatherResponseNoAlerts(t *testing.T) {
	// Test weather data without alerts or UV
	weatherData := &OneCallResponse{
		Current: CurrentWeather{
			Temp:      15.0,
			FeelsLike: 14.5,
			WindSpeed: 2.1,
			Humidity:  70,
			Pressure:  1020,
			Clouds:    20,
			UVI:       0, // No UV index
			Weather: []WeatherCondition{{
				Description: "clear sky",
			}},
		},
		Alerts: []Alert{}, // No alerts
	}

	result := formatWeatherResponse("London", "GB", weatherData)

	// Should not contain UV index or alerts
	if strings.Contains(result, "UV index") {
		t.Errorf("Result should not contain UV index when UVI is 0: %s", result)
	}

	if strings.Contains(result, "alert") {
		t.Errorf("Result should not contain alerts when none present: %s", result)
	}

	expectedParts := []string{
		"London, GB",
		"Temperature: 15.0°C",
		"feels like: 14.5°C",
		"clear sky",
	}

	for _, part := range expectedParts {
		if !strings.Contains(result, part) {
			t.Errorf("Expected result to contain '%s', but got: %s", part, result)
		}
	}
}

func TestGetCoordinatesMockServer(t *testing.T) {
	// Mock geocoding API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "Helsinki") {
			response := GeocodingResponse{{
				Name:    "Helsinki",
				Lat:     60.1699,
				Lon:     24.9384,
				Country: "FI",
			}}
			_ = json.NewEncoder(w).Encode(response)
			return
		}
		if strings.Contains(r.URL.String(), "UnknownCity") {
			_ = json.NewEncoder(w).Encode(GeocodingResponse{})
			return
		}
		http.Error(w, "Not found", 404)
	}))
	defer server.Close()

	// Test function that uses our mock server
	testGetCoords := func(appid, location string) (float64, float64, string, string, error) {
		geoURL := fmt.Sprintf("%s?q=%s&limit=1&appid=%s", server.URL, location, appid)
		resp, err := http.Get(geoURL)
		if err != nil {
			return 0, 0, "", "", fmt.Errorf("geocoding API request failed: %v", err)
		}
		defer resp.Body.Close()

		var geoData GeocodingResponse
		err = json.NewDecoder(resp.Body).Decode(&geoData)
		if err != nil {
			return 0, 0, "", "", fmt.Errorf("failed to parse geocoding response: %v", err)
		}

		if len(geoData) == 0 {
			return 0, 0, "", "", fmt.Errorf("location not found: %s", location)
		}

		loc := geoData[0]
		return loc.Lat, loc.Lon, loc.Name, loc.Country, nil
	}

	// Test valid location
	lat, lon, name, country, err := testGetCoords("test-api-key", "Helsinki")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if lat != 60.1699 || lon != 24.9384 || name != "Helsinki" || country != "FI" {
		t.Errorf("Expected Helsinki coordinates, got lat=%f lon=%f name=%s country=%s",
			lat, lon, name, country)
	}

	// Test unknown location
	_, _, _, _, err = testGetCoords("test-api-key", "UnknownCity")
	if err == nil {
		t.Error("Expected error for unknown location")
	}
}

func TestOpenWeatherMissingAPIKey(t *testing.T) {
	// Save original env var
	originalKey := os.Getenv("OPENWEATHERMAP_API_KEY")

	// Unset the API key
	os.Unsetenv("OPENWEATHERMAP_API_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("OPENWEATHERMAP_API_KEY", originalKey)
		}
	}()

	_, err := OpenWeather("Helsinki")
	if err == nil {
		t.Error("Expected error when API key is missing")
	}

	if !strings.Contains(err.Error(), "OPENWEATHERMAP_API_KEY environment variable not set") {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}

func TestOpenWeatherDefaultLocation(t *testing.T) {
	// This test only verifies the default location behavior without making API calls
	// We can't easily test the full function without mocking HTTP calls

	// Just test that empty args defaults to Helsinki
	// The function would need API key and actual HTTP calls to work fully
	originalKey := os.Getenv("OPENWEATHERMAP_API_KEY")
	os.Unsetenv("OPENWEATHERMAP_API_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("OPENWEATHERMAP_API_KEY", originalKey)
		}
	}()

	_, err := OpenWeather("")
	// Should get API key error, not a location parsing error
	if err == nil || !strings.Contains(err.Error(), "OPENWEATHERMAP_API_KEY") {
		t.Errorf("Expected API key error, got: %v", err)
	}
}

func TestFormatWeatherResponseMultipleAlerts(t *testing.T) {
	// Test weather data with multiple alerts
	weatherData := &OneCallResponse{
		Current: CurrentWeather{
			Temp:      25.0,
			FeelsLike: 27.0,
			WindSpeed: 8.5,
			Humidity:  45,
			Pressure:  1005,
			Clouds:    90,
			UVI:       9.1,
			Weather: []WeatherCondition{{
				Description: "thunderstorm",
			}},
		},
		Alerts: []Alert{
			{Event: "Severe Thunderstorm Warning"},
			{Event: "Flash Flood Watch"},
			{Event: "High Wind Advisory"},
		},
	}

	result := formatWeatherResponse("Dallas", "US", weatherData)

	// Should show first alert and count of additional alerts
	expectedParts := []string{
		"Dallas, US",
		"Temperature: 25.0°C",
		"thunderstorm",
		"UV index: 9.1",
		"⚠️ Severe Thunderstorm Warning (+2 more alerts)",
	}

	for _, part := range expectedParts {
		if !strings.Contains(result, part) {
			t.Errorf("Expected result to contain '%s', but got: %s", part, result)
		}
	}
}
