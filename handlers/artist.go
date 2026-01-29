package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"groupie-tracker/models"
)

const baseAPI = "https://groupietrackers.herokuapp.com/api"

var templates = template.Must(
	template.ParseGlob("templates/*.html"),
)

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
	artists, err := GetArtists()
	if err != nil {
		http.Error(w, "Impossible de récupérer les artistes", http.StatusInternalServerError)
		return
	}
	err = templates.ExecuteTemplate(w, "layout.html", artists)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func ArtistDetail(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/artist/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	artists, err := GetArtists()

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
	locations, err := GetLocationsByID(id)

	data := struct {
		Artist    *models.Artist
		Dates     []string
		Locations []string
	}{
		found,
		dates,
		locations,
	}

	err = templates.ExecuteTemplate(w, "layout.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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
