package command

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/lepinkainen/lambdabot/lambda"
)

// GeocodingResponse represents the response from OpenWeatherMap Geocoding API
type GeocodingResponse []struct {
	Name    string  `json:"name"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
	Country string  `json:"country"`
	State   string  `json:"state,omitempty"`
}

// OneCallResponse represents the response from OpenWeatherMap One Call API 3.0
type OneCallResponse struct {
	Lat      float64        `json:"lat"`
	Lon      float64        `json:"lon"`
	Timezone string         `json:"timezone"`
	Current  CurrentWeather `json:"current"`
	Alerts   []Alert        `json:"alerts,omitempty"`
}

// CurrentWeather represents current weather data from One Call API 3.0
type CurrentWeather struct {
	Dt         int64              `json:"dt"`
	Sunrise    int64              `json:"sunrise"`
	Sunset     int64              `json:"sunset"`
	Temp       float64            `json:"temp"`
	FeelsLike  float64            `json:"feels_like"`
	Pressure   int                `json:"pressure"`
	Humidity   int                `json:"humidity"`
	DewPoint   float64            `json:"dew_point"`
	UVI        float64            `json:"uvi"`
	Clouds     int                `json:"clouds"`
	Visibility int                `json:"visibility"`
	WindSpeed  float64            `json:"wind_speed"`
	WindDeg    int                `json:"wind_deg"`
	WindGust   float64            `json:"wind_gust,omitempty"`
	Weather    []WeatherCondition `json:"weather"`
}

// WeatherCondition represents weather condition data
type WeatherCondition struct {
	ID          int    `json:"id"`
	Main        string `json:"main"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

// Alert represents weather alerts
type Alert struct {
	SenderName  string `json:"sender_name"`
	Event       string `json:"event"`
	Start       int64  `json:"start"`
	End         int64  `json:"end"`
	Description string `json:"description"`
}

// OpenWeather command handler using One Call API 3.0
func OpenWeather(args string) (string, error) {
	if args == "" {
		args = "Helsinki"
	}

	appid := os.Getenv("OPENWEATHERMAP_API_KEY")
	if appid == "" {
		return "", fmt.Errorf("OPENWEATHERMAP_API_KEY environment variable not set")
	}

	// Get coordinates for the location
	lat, lon, locationName, country, err := getCoordinates(appid, args)
	if err != nil {
		return "", fmt.Errorf("unable to geocode location %s: %v", args, err)
	}

	// Get weather data from One Call API
	weatherData, err := getOneCallWeather(appid, lat, lon)
	if err != nil {
		return "", fmt.Errorf("unable to get weather data: %v", err)
	}

	return formatWeatherResponse(locationName, country, weatherData), nil
}

// getCoordinates gets latitude and longitude for a location
func getCoordinates(appid, location string) (lat, lon float64, name, country string, err error) {
	// Geocoding API call with URL encoding for location
	geoURL := fmt.Sprintf("http://api.openweathermap.org/geo/1.0/direct?q=%s&limit=1&appid=%s", url.QueryEscape(location), appid)
	resp, err := http.Get(geoURL)
	if err != nil {
		return 0, 0, "", "", fmt.Errorf("geocoding API request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, "", "", fmt.Errorf("failed to read geocoding response: %v", err)
	}

	var geoData GeocodingResponse
	err = json.Unmarshal(body, &geoData)
	if err != nil {
		return 0, 0, "", "", fmt.Errorf("failed to parse geocoding response: %v", err)
	}

	if len(geoData) == 0 {
		return 0, 0, "", "", fmt.Errorf("location not found: %s", location)
	}

	loc := geoData[0]
	return loc.Lat, loc.Lon, loc.Name, loc.Country, nil
}

// getOneCallWeather fetches weather data from One Call API 3.0
func getOneCallWeather(appid string, lat, lon float64) (*OneCallResponse, error) {
	apiURL := fmt.Sprintf("https://api.openweathermap.org/data/3.0/onecall?lat=%s&lon=%s&appid=%s&units=metric&exclude=minutely,hourly,daily",
		strconv.FormatFloat(lat, 'f', 6, 64),
		strconv.FormatFloat(lon, 'f', 6, 64),
		appid)

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("One Call API request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("One Call API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read One Call API response: %v", err)
	}

	var weatherData OneCallResponse
	err = json.Unmarshal(body, &weatherData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse One Call API response: %v", err)
	}

	return &weatherData, nil
}

// formatWeatherResponse formats the weather data into a human-readable string
func formatWeatherResponse(locationName, country string, weatherData *OneCallResponse) string {
	current := weatherData.Current
	description := ""
	if len(current.Weather) > 0 {
		description = current.Weather[0].Description
	}

	// Build base response
	result := fmt.Sprintf("%s, %s: Temperature: %.1f°C, feels like: %.1f°C, wind: %.1f m/s, humidity: %d%%, pressure: %dhPa, cloudiness: %d%%, %s",
		locationName, country, current.Temp, current.FeelsLike, current.WindSpeed, current.Humidity, current.Pressure, current.Clouds, description)

	// Add UV index if available
	if current.UVI > 0 {
		result += fmt.Sprintf(", UV index: %.1f", current.UVI)
	}

	// Add weather alerts if any
	if len(weatherData.Alerts) > 0 {
		if len(weatherData.Alerts) == 1 {
			result += fmt.Sprintf(" ⚠️ %s", weatherData.Alerts[0].Event)
		} else {
			// Multiple alerts - show count and first event
			result += fmt.Sprintf(" ⚠️ %s (+%d more alerts)", weatherData.Alerts[0].Event, len(weatherData.Alerts)-1)
		}
	}

	return result
}

func init() {
	lambda.RegisterHandler("weather", OpenWeather)
	lambda.RegisterHandler("forecast", OpenWeather)
}

// TODO: Enhanced forecasts using One Call API 3.0 hourly/daily data
// TODO: Weather alert notifications
// TODO: Historical weather data
