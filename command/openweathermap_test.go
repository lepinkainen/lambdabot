package command

import (
	"io/ioutil"
	"testing"
)

func loadJSON(name string) []byte {
	bytes, err := ioutil.ReadFile(name)
	if err != nil {
		panic(err)
	}

	return bytes
}

func Test_parseWeather(t *testing.T) {
	type args struct {
		bytes []byte
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"GoldenFile1", args{bytes: loadJSON("weather.json")}, "Helsinki, FI: Temperature: 14.5°C, feels like: 13.9°C, wind: 3.1 m/s, humidity: 58%, pressure: 1019hPa, cloudiness: 100%", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseWeather(tt.args.bytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseWeather() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseWeather() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_feelsLike(t *testing.T) {
	type args struct {
		temperature float64
		wind        float64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		// these values are confirmed using online calculators with the same input
		{"WindChill", args{temperature: 2, wind: 3.6}, -1.6},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := feelsLike(tt.args.temperature, tt.args.wind); got != tt.want {
				t.Errorf("feelsLike() = %v, want %v", got, tt.want)
			}
		})
	}
}
