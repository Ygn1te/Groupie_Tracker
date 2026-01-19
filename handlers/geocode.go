package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

// Fallback : charge aussi le .env directement dans handlers
// => comme ça, même si main oublie de le faire, ça marche.
func init() {
	godotenv.Load()
}

type GeoLocation struct {
	Name string  `json:"name"`
	Lat  float64 `json:"lat"`
	Lng  float64 `json:"lng"`
}

type openCageResponse struct {
	Results []struct {
		Geometry struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		} `json:"geometry"`
		Formatted string `json:"formatted"`
	} `json:"results"`
}

var (
	geoCache = map[string]GeoLocation{}
	geoMutex sync.RWMutex
)

func normalizeLocation(loc string) string {
	// "paris-france" -> "paris, france"
	loc = strings.ReplaceAll(loc, "_", " ")
	loc = strings.ReplaceAll(loc, "-", ", ")
	return loc
}

func GeocodeOpenCage(location string) (GeoLocation, error) {
	// Cache (évite de bouffer le quota)
	geoMutex.RLock()
	if val, ok := geoCache[location]; ok {
		geoMutex.RUnlock()
		return val, nil
	}
	geoMutex.RUnlock()

	key := os.Getenv("OPENCAGE_KEY")
	if key == "" {
		return GeoLocation{}, fmt.Errorf("missing OPENCAGE_KEY env var")
	}

	q := normalizeLocation(location)

	endpoint := fmt.Sprintf(
		"https://api.opencagedata.com/geocode/v1/json?q=%s&key=%s&limit=1",
		url.QueryEscape(q),
		key,
	)

	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Get(endpoint)
	if err != nil {
		return GeoLocation{}, err
	}
	defer resp.Body.Close()

	var data openCageResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return GeoLocation{}, err
	}

	if len(data.Results) == 0 {
		return GeoLocation{}, fmt.Errorf("no results for %s", location)
	}

	result := GeoLocation{
		Name: data.Results[0].Formatted,
		Lat:  data.Results[0].Geometry.Lat,
		Lng:  data.Results[0].Geometry.Lng,
	}

	geoMutex.Lock()
	geoCache[location] = result
	geoMutex.Unlock()

	return result, nil
}
