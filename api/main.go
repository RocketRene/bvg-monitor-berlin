package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

const (
	pfahlerstraseURL  = "https://v6.bvg.transport.rest/stops/900086159/departures?duration=100&results=10&linesOfStops=false&remarks=false&language=en"
	karlbonnhoeferURL = "https://v6.bvg.transport.rest/stops/900096458/departures?direction=900079221&duration=100&linesOfStops=false&remarks=true&language=en"
	rosDirection      = "Rosenthal Nord"
	jungfDirection    = "S+U Jungfernheide"
	timezone          = "Europe/Berlin"
	wheatherURL       = "https://api.open-meteo.com/v1/forecast?latitude=52.52&longitude=13.41&daily=sunset,precipitation_probability_max&current_weather=true&timezone=Europe%2FBerlin"
)

type Wheather struct {
	Latitude             float64 `json:"latitude"`
	Longitude            float64 `json:"longitude"`
	GenerationtimeMs     float64 `json:"generationtime_ms"`
	UtcOffsetSeconds     int     `json:"utc_offset_seconds"`
	Timezone             string  `json:"timezone"`
	TimezoneAbbreviation string  `json:"timezone_abbreviation"`
	Elevation            float64 `json:"elevation"`
	CurrentWeather       struct {
		Temperature   float64 `json:"temperature"`
		Windspeed     float64 `json:"windspeed"`
		Winddirection float64 `json:"winddirection"`
		Weathercode   int     `json:"weathercode"`
		IsDay         int     `json:"is_day"`
		Time          string  `json:"time"`
	} `json:"current_weather"`
	DailyUnits struct {
		Time                        string `json:"time"`
		Sunset                      string `json:"sunset"`
		PrecipitationProbabilityMax string `json:"precipitation_probability_max"`
	} `json:"daily_units"`
	Daily struct {
		Time                        []string `json:"time"`
		Sunset                      []string `json:"sunset"`
		PrecipitationProbabilityMax []int    `json:"precipitation_probability_max"`
	} `json:"daily"`
}

type Bus struct {
	Direction string `json:"direction"`
	When      string `json:"when"`
}

func main() {
	// Load location for timezone
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		log.Fatal(err)
	}

	// Create HTTP handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Update boxes with timings
		client := http.DefaultClient
		resp, err := client.Get(pfahlerstraseURL)

		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		var data struct {
			Departures []Bus `json:"departures"`
		}
		err = json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		rosentahlRemaining, jungfRemaining := []int{}, []int{}
		now := time.Now().In(loc)
		for _, bus := range data.Departures {
			when, err := time.ParseInLocation(time.RFC3339, bus.When, time.UTC)
			if err != nil {
				log.Println(err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			remaining := int(when.Sub(now).Minutes())
			switch bus.Direction {
			case rosDirection:
				rosentahlRemaining = append(rosentahlRemaining, remaining)
			case jungfDirection:
				jungfRemaining = append(jungfRemaining, remaining)
			}
		}
		//only show the next 3 busses
		if len(rosentahlRemaining) > 3 {
			rosentahlRemaining = rosentahlRemaining[:3]
		}
		if len(jungfRemaining) > 3 {
			jungfRemaining = jungfRemaining[:3]
		}

		// add karlbonnhoefer
		resp, err = client.Get(karlbonnhoeferURL)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		err = json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		u8remaining := []int{}
		for _, bus := range data.Departures {
			when, err := time.ParseInLocation(time.RFC3339, bus.When, time.UTC)
			if err != nil {
				log.Println(err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			remaining := int(when.Sub(now).Minutes())
			u8remaining = append(u8remaining, remaining)
		}
		// delete everything under 10 minutes and only show the next 3 u8 departures
		for i := len(u8remaining) - 1; i >= 0; i-- {
			if u8remaining[i] < 10 {
				u8remaining = append(u8remaining[:i], u8remaining[i+1:]...)
			}
		}
		if len(u8remaining) > 3 {
			u8remaining = u8remaining[:3]
		}

		// add wheather
		resp, err = client.Get(wheatherURL)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		var wheather Wheather
		err = json.NewDecoder(resp.Body).Decode(&wheather)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		wheatherResponse := map[string]interface{}{
			"temperature":                   wheather.CurrentWeather.Temperature,
			"precipitation_probability_max": wheather.Daily.PrecipitationProbabilityMax[0],
			"sunset":                        wheather.Daily.Sunset[0][11:16],
		}

		//Todo add U 8 line

		// Write response as JSON
		busResponse := map[string][]int{
			"Rosenthal":  rosentahlRemaining,
			"Jungfernh.": jungfRemaining,
			"U8":         u8remaining,
		}

		// merge wheather and Busresponse
		response := map[string]interface{}{
			"bus":       busResponse,
			"wheather":  wheatherResponse,
			"timestamp": time.Now().Format("15:04"),
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})

	// Start server
	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
	log.Println("Starting server on :8080...")
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
