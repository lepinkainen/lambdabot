package command

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"

	"github.com/lepinkainen/lambdabot/lambda"
	log "github.com/sirupsen/logrus"
)

type OpenWeatherMapJSON struct {
	Base   string `json:"base"`
	Clouds struct {
		All int `json:"all"`
	} `json:"clouds"`
	Cod   int `json:"cod"`
	Coord struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	} `json:"coord"`
	Dt   int `json:"dt"`
	ID   int `json:"id"`
	Main struct {
		Humidity int     `json:"humidity"`
		Pressure int     `json:"pressure"`
		Temp     float64 `json:"temp"`
		TempMax  int     `json:"temp_max"`
		TempMin  float64 `json:"temp_min"`
	} `json:"main"`
	Name string `json:"name"`
	Sys  struct {
		Country string  `json:"country"`
		ID      int     `json:"id"`
		Message float64 `json:"message"`
		Sunrise int     `json:"sunrise"`
		Sunset  int     `json:"sunset"`
		Type    int     `json:"type"`
	} `json:"sys"`
	Visibility int `json:"visibility"`
	Weather    []struct {
		Description string `json:"description"`
		Icon        string `json:"icon"`
		ID          int    `json:"id"`
		Main        string `json:"main"`
	} `json:"weather"`
	Wind struct {
		Deg   int     `json:"deg"`
		Speed float64 `json:"speed"`
	} `json:"wind"`
}

func OpenWeather(args string) (string, error) {

	if args == "" {
		args = "Helsinki"
	}

	appid := os.Getenv("OPENWEATHERMAP_API_KEY")

	apiurl := fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?appid=%s&units=metric&q=%s", appid, args)

	res, err := http.Get(apiurl)
	if err != nil {
		log.Errorf("Unable to get API response: %v", err)
		return "", err
	}
	defer res.Body.Close()

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Errorf("Unable to read response: %v", err)
		return "", err
	}

	response, err := parseWeather(bytes)

	return response, err
}

// Use a formula to calculate wind chill from current temperature in Celsius and wind in m/s
// returns a float with one decimal place
func feelsLike(temperature, wind float64) float64 {

	windExp := math.Pow(wind, 0.16)

	feelsLike := 13.12 + 0.6215*temperature - 13.956*windExp + 0.4867*temperature*windExp

	return math.Round(feelsLike*10) / 10
}

func parseWeather(bytes []byte) (string, error) {
	data := &OpenWeatherMapJSON{}

	err := json.Unmarshal(bytes, &data)
	if err != nil {
		log.Errorf("Unable to unmarshal result JSON")
		return "", err
	}

	if data.Cod != 200 {
		return "", fmt.Errorf("API Error: %v", data)
	}

	feelsLike := feelsLike(float64(data.Main.Temp), float64(data.Wind.Speed))
	result := fmt.Sprintf("%s, %s: Temperature: %.1f°C, feels like: %.1f°C, wind: %.1f m/s, humidity: %d%%, pressure: %dhPa, cloudiness: %d%%",
		data.Name, data.Sys.Country, data.Main.Temp, feelsLike, data.Wind.Speed, data.Main.Humidity, data.Main.Pressure, data.Clouds.All)

	return result, nil
}

func init() {
	lambda.RegisterHandler("weather", OpenWeather)
	lambda.RegisterHandler("forecast", OpenWeather)
}

// TODO: Forecasts
// https://openweathermap.org/forecast5
