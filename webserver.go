package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

import "github.com/evizitei/weatherman"

func main() {
	http.HandleFunc("/", hello)
	http.HandleFunc("/weather/", weather)
	http.ListenAndServe(":9111", nil)
}

func hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello!"))
}

func weather(w http.ResponseWriter, r *http.Request) {
	city := strings.SplitN(r.URL.Path, "/", 3)[2]

	data, err := weatherQuery(city)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(data)
}

func weatherQuery(city string) (weatherman.WeatherData, error) {
	mw := weatherman.WeatherAggregator{
		weatherman.OpenWeatherMap{},
		weatherman.WeatherUnderground{ApiKey: "aa7f411e26da74aa"},
	}

	begin := time.Now()
	temp, err := mw.Temperature(city)
	if err != nil {
		return weatherman.WeatherData{}, err
	}

	d := weatherman.WeatherData{
		City: city,
		Temp: temp,
		Took: time.Since(begin).String(),
	}

	return d, nil
}
