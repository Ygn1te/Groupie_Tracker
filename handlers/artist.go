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

	tmpl.Execute(w, artists)
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

	data := struct {
		Artist    *models.Artist
		Dates     []string
		Locations []string
	}{found, dates, locations}

	tmpl, err := template.ParseFiles("templates/artist.html")
	if err != nil {
		http.Error(w, "Erreur template", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, data)
}

func Search(w http.ResponseWriter, r *http.Request) {
	q := strings.ToLower(r.URL.Query().Get("q"))
	artists, err := GetArtists()
	if err != nil {
		http.Error(w, "Impossible de récupérer artistes", http.StatusInternalServerError)
		return
	}

	var res []models.Artist
	for _, a := range artists {
		if strings.Contains(strings.ToLower(a.Name), q) {
			res = append(res, a)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
