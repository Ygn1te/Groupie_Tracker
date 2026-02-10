package handlers

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"groupie-tracker/models"
)

const baseAPI = "https://groupietrackers.herokuapp.com/api"

func GetArtists() ([]models.Artist, error) {
	resp, err := http.Get(baseAPI + "/artists")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var artists []models.Artist
	err = json.NewDecoder(resp.Body).Decode(&artists)
	return artists, err
}

func GetDatesByID(id int) ([]string, error) {
	resp, err := http.Get(baseAPI + "/dates/" + strconv.Itoa(id))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var d models.DatesResp
	err = json.NewDecoder(resp.Body).Decode(&d)
	return d.Dates, err
}

func GetLocationsByID(id int) ([]string, error) {
	resp, err := http.Get(baseAPI + "/locations/" + strconv.Itoa(id))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var l models.LocationsResp
	err = json.NewDecoder(resp.Body).Decode(&l)
	return l.Locations, err
}

// Données envoyées au template index (liste + recherche)
type IndexData struct {
	Artists []models.Artist
	Query   string
}

func Home(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "Erreur template", http.StatusInternalServerError)
		return
	}

	artists, err := GetArtists()
	if err != nil {
		http.Error(w, "Impossible de récupérer les artistes", http.StatusInternalServerError)
		return
	}

	data := IndexData{
		Artists: artists,
		Query:   "",
	}

	tmpl.Execute(w, data)
}

func ArtistDetail(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/artist/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	artists, err := GetArtists()
	if err != nil {
		http.Error(w, "Impossible de lire artistes", http.StatusInternalServerError)
		return
	}

	var found *models.Artist
	for _, a := range artists {
		if a.ID == id {
			art := a
			found = &art
			break
		}
	}
	if found == nil {
		http.NotFound(w, r)
		return
	}

	dates, err := GetDatesByID(id)
	if err != nil {
		log.Println("Dates error:", err)
		dates = []string{}
	}

	locations, err := GetLocationsByID(id)
	if err != nil {
		log.Println("Locations error:", err)
		locations = []string{}
	}

	// Géocodage (OpenCage) + cache
	var geoLocations []GeoLocation
	for _, loc := range locations {
		gl, err := GeocodeOpenCage(loc)
		if err == nil {
			geoLocations = append(geoLocations, gl)
		} else {
			log.Println("Geocode error:", err)
		}
	}

	// On injecte du JSON directement pour le JS (plus fiable)
	geoJSONBytes, err := json.Marshal(geoLocations)
	if err != nil {
		log.Println("GeoJSON error:", err)
		geoJSONBytes = []byte("[]")
	}

	data := struct {
		Artist       *models.Artist
		Dates        []string
		Locations    []string
		GeoLocations []GeoLocation
		GeoJSON      template.JS
	}{
		Artist:       found,
		Dates:        dates,
		Locations:    locations,
		GeoLocations: geoLocations,
		GeoJSON:      template.JS(geoJSONBytes),
	}

	tmpl, err := template.ParseFiles("templates/artist.html")
	if err != nil {
		http.Error(w, "Erreur template", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, data)
}

func Search(w http.ResponseWriter, r *http.Request) {
	raw := strings.TrimSpace(r.URL.Query().Get("q"))
	q := strings.ToLower(raw)

	artists, err := GetArtists()
	if err != nil {
		http.Error(w, "Impossible de récupérer artistes", http.StatusInternalServerError)
		return
	}

	// Recherche vide -> on affiche tout
	if q == "" {
		tmpl, err := template.ParseFiles("templates/index.html")
		if err != nil {
			http.Error(w, "Erreur template", http.StatusInternalServerError)
			return
		}

		data := IndexData{Artists: artists, Query: ""}
		tmpl.Execute(w, data)
		return
	}

	// ✅ NAME ONLY + ✅ COMMENCE PAR (prefix)
	var res []models.Artist
	for _, a := range artists {
		if strings.HasPrefix(strings.ToLower(a.Name), q) {
			res = append(res, a)
		}
	}

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "Erreur template", http.StatusInternalServerError)
		return
	}

	data := IndexData{
		Artists: res,
		Query:   raw,
	}

	tmpl.Execute(w, data)
}

// ✅ Suggestions JSON pour l'autocomplete (name commence par)
func Suggest(w http.ResponseWriter, r *http.Request) {
	raw := strings.TrimSpace(r.URL.Query().Get("q"))
	q := strings.ToLower(raw)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if q == "" {
		json.NewEncoder(w).Encode([]any{})
		return
	}

	artists, err := GetArtists()
	if err != nil {
		http.Error(w, "Impossible de récupérer artistes", http.StatusInternalServerError)
		return
	}

	type SuggestItem struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	const limit = 10
	out := make([]SuggestItem, 0, limit)

	for _, a := range artists {
		if strings.HasPrefix(strings.ToLower(a.Name), q) {
			out = append(out, SuggestItem{ID: a.ID, Name: a.Name})
			if len(out) >= limit {
				break
			}
		}
	}

	json.NewEncoder(w).Encode(out)
}
