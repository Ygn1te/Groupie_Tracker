package handlers

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"groupie_tracker/models"
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
	tmpl, err := template.ParseFiles("templates/band_list.html")
	if err != nil {
		http.Error(w, "Erreur template", http.StatusInternalServerError)
		return
	}

	// Parse query params for search and filters
	q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))

	// range filters (capture raw strings to preserve values in the UI)
	creationMinStr := r.URL.Query().Get("creation_min")
	creationMaxStr := r.URL.Query().Get("creation_max")
	albumMinStr := r.URL.Query().Get("album_min")
	albumMaxStr := r.URL.Query().Get("album_max")
	membersMinStr := r.URL.Query().Get("members_min")
	membersMaxStr := r.URL.Query().Get("members_max")

	creationMin, _ := strconv.Atoi(creationMinStr)
	creationMax, _ := strconv.Atoi(creationMaxStr)
	albumMin, _ := strconv.Atoi(albumMinStr)
	albumMax, _ := strconv.Atoi(albumMaxStr)
	membersMin, _ := strconv.Atoi(membersMinStr)
	membersMax, _ := strconv.Atoi(membersMaxStr)

	// checkbox filters: locations - may be multiple
	selectedLocations := r.URL.Query()["location"]
	selMap := map[string]bool{}
	for _, s := range selectedLocations {
		selMap[s] = true
	}

	artists, err := GetArtists()
	if err != nil {
		http.Error(w, "Impossible de récupérer les artistes", http.StatusInternalServerError)
		return
	}

	// Build a set of available locations by fetching each artist's locations
	locSet := map[string]struct{}{}
	artistLocations := make(map[int][]string)
	for _, a := range artists {
		locs, err := GetLocationsByID(a.ID)
		if err != nil {
			// Log error and continue; don't fail the whole page
			log.Println("GetLocationsByID error for id", a.ID, err)
			continue
		}
		artistLocations[a.ID] = locs
		for _, l := range locs {
			// normalize: treat parent regions as available too (e.g., "Seattle, Washington, USA" -> "Washington, USA")
			locSet[l] = struct{}{}
			parts := strings.Split(l, ",")
			if len(parts) >= 2 {
				parent := strings.TrimSpace(strings.Join(parts[len(parts)-2:], ","))
				locSet[parent] = struct{}{}
			}
		}
	}

	var availableLocations []string
	for l := range locSet {
		availableLocations = append(availableLocations, l)
	}

	// Apply filters
	var filtered []models.Artist
	for _, a := range artists {
		// search by name (existing search bar behavior)
		if q != "" {
			if !strings.Contains(strings.ToLower(a.Name), q) {
				continue
			}
		}

		// creation date range
		if creationMin != 0 && a.CreationDate < creationMin {
			continue
		}
		if creationMax != 0 && a.CreationDate > creationMax {
			continue
		}

		// members count range
		mcount := len(a.Members)
		if membersMin != 0 && mcount < membersMin {
			continue
		}
		if membersMax != 0 && mcount > membersMax {
			continue
		}

		// first album year
		year := parseYear(a.FirstAlbum)
		if albumMin != 0 && year != 0 && year < albumMin {
			continue
		}
		if albumMax != 0 && year != 0 && year > albumMax {
			continue
		}

		// locations (checkboxes): if any selected, artist must have at least one match
		if len(selectedLocations) > 0 {
			locs := artistLocations[a.ID]
			if !artistMatchesLocations(locs, selectedLocations) {
				continue
			}
		}

		filtered = append(filtered, a)
	}

	data := struct {
		Query              string
		Artists            []models.Artist
		LocationsAvailable []string
		SelectedLocations  []string
		SelectedLocationsMap map[string]bool
		CreationMin string
		CreationMax string
		AlbumMin string
		AlbumMax string
		MembersMin string
		MembersMax string
	}{
		Query:              q,
		Artists:            filtered,
		LocationsAvailable: availableLocations,
		SelectedLocations:  selectedLocations,
		SelectedLocationsMap: selMap,
		CreationMin: creationMinStr,
		CreationMax: creationMaxStr,
		AlbumMin: albumMinStr,
		AlbumMax: albumMaxStr,
		MembersMin: membersMinStr,
		MembersMax: membersMaxStr,
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Erreur rendering", http.StatusInternalServerError)
		return
	}
}

// parseYear attempts to extract a 4-digit year from a string; returns 0 if not found
func parseYear(s string) int {
	// look for 4-digit sequence
	for i := 0; i+4 <= len(s); i++ {
		sub := s[i : i+4]
		if y, err := strconv.Atoi(sub); err == nil && y >= 1000 && y <= 3000 {
			return y
		}
	}
	return 0
}

// artistMatchesLocations returns true if any artist location matches any selected location.
// Matching is permissive: exact match or parent region match (e.g., "Seattle, Washington, USA" matches "Washington, USA").
func artistMatchesLocations(artistLocs, selected []string) bool {
	selSet := map[string]struct{}{}
	for _, s := range selected {
		selSet[strings.ToLower(strings.TrimSpace(s))] = struct{}{}
	}
	for _, al := range artistLocs {
		al = strings.ToLower(strings.TrimSpace(al))
		if _, ok := selSet[al]; ok {
			return true
		}
		parts := strings.Split(al, ",")
		if len(parts) >= 2 {
			parent := strings.ToLower(strings.TrimSpace(strings.Join(parts[len(parts)-2:], ",")))
			if _, ok := selSet[parent]; ok {
				return true
			}
		}
		if len(parts) >= 1 {
			country := strings.ToLower(strings.TrimSpace(parts[len(parts)-1]))
			if _, ok := selSet[country]; ok {
				return true
			}
		}
	}
	return false
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

	tmpl, err := template.ParseFiles("templates/band_info.html")
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
