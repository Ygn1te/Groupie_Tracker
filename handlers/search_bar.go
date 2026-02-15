package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"groupie_tracker/models"
)

func SearchArtists(w http.ResponseWriter, r *http.Request) {
	q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))

	artists, err := GetArtists()
	if err != nil {
		http.Error(w, "Impossible de récupérer les artistes", http.StatusInternalServerError)
		return
	}

	filtered := make([]models.Artist, 0)
	for _, a := range artists {
		name := strings.ToLower(strings.TrimSpace(a.Name))
		if q == "" || strings.HasPrefix(name, q) {
			filtered = append(filtered, a)
		}
	}

	if len(filtered) > 5 {
		filtered = filtered[:5]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(filtered)
}

// Version serveur du filtrage (optionnelle)
// Peut être utilisée si tu veux filtrer AVANT d'afficher la page HTML

func FilterArtists(q string) ([]models.Artist, error) {
	artists, err := GetArtists()
	if err != nil {
		return nil, err
	}

	q = strings.ToLower(strings.TrimSpace(q))
	var filtered []models.Artist

	for _, a := range artists {
		name := strings.ToLower(strings.TrimSpace(a.Name))
		if q == "" || strings.HasPrefix(name, q) {
			filtered = append(filtered, a)
		}
	}

	if len(filtered) > 5 {
		filtered = filtered[:5]
	}

	return filtered, nil
}
