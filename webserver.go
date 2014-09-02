package webserver

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

type weatherData struct {
	City string  `json:"city"`
	Temp float64 `json:"temp"`
	Took string  `json:"took"`
}

type weatherProvider interface {
	temperature(city string) (float64, error) //in Kelvin
}

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

func weatherQuery(city string) (weatherData, error) {
	mw := weatherAggregator{
		openWeatherMap{},
		weatherUnderground{apiKey: "aa7f411e26da74aa"},
	}

	begin := time.Now()
	temp, err := mw.temperature(city)
	if err != nil {
		return weatherData{}, err
	}

	d := weatherData{
		City: city,
		Temp: temp,
		Took: time.Since(begin).String(),
	}

	return d, nil
}

type weatherAggregator []weatherProvider

func (w weatherAggregator) temperature(city string) (float64, error) {
	sum := 0.0
	for _, provider := range w {
		k, err := provider.temperature(city)
		if err != nil {
			return 0, err
		}
		sum += k
	}

	return sum / float64(len(w)), nil
}

type openWeatherMap struct{}

func (w openWeatherMap) temperature(city string) (float64, error) {
	endpoint := "http://api.openweathermap.org/data/2.5/weather?q="
	resp, err := http.Get(endpoint + city)

	if err != nil {
		return 0, err
	}

	defer resp.Body.Close()

	var d struct {
		Main struct {
			Kelvin float64 `json:"temp"`
		} `json:"main"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return 0, err
	}

	log.Printf("openWeatherMap %s: %.2f", city, d.Main.Kelvin)
	return d.Main.Kelvin, nil
}

type weatherUnderground struct {
	apiKey string
}

func (w weatherUnderground) temperature(city string) (float64, error) {
	endpoint := "http://api.wunderground.com/api/" + w.apiKey + "/conditions/q/"
	resp, err := http.Get(endpoint + city + ".json")
	if err != nil {
		return 0, err
	}

	defer resp.Body.Close()

	var d struct {
		Observation struct {
			Celsius float64 `json:"temp_c"`
		} `json:"current_observation"`
	}

	kelvin := d.Observation.Celsius + 273.15
	log.Printf("weatherUnderground: %s: %.2f", city, kelvin)
	return kelvin, nil
}
