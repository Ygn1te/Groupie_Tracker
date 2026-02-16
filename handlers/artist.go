package handlers

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

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

type locationsIndexResp struct {
	Index []struct {
		ID        int      `json:"id"`
		Locations []string `json:"locations"`
	} `json:"index"`
}

func GetAllLocations() (map[int][]string, error) {
	resp, err := http.Get(baseAPI + "/locations")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var l locationsIndexResp
	if err := json.NewDecoder(resp.Body).Decode(&l); err != nil {
		return nil, err
	}

	out := make(map[int][]string, len(l.Index))
	for _, it := range l.Index {
		out[it.ID] = it.Locations
	}
	return out, nil
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

type cachedData struct {
	Artists   []models.Artist
	Locations map[int][]string
	LocOpts   []string
	FetchedAt time.Time
}

var (
	cacheMu sync.Mutex
	cache   cachedData
)

func getCachedData() (cachedData, error) {
	cacheMu.Lock()
	defer cacheMu.Unlock()

	if time.Since(cache.FetchedAt) < 10*time.Minute && cache.FetchedAt != (time.Time{}) {
		return cache, nil
	}

	artists, err := GetArtists()
	if err != nil {
		return cachedData{}, err
	}

	locMap, err := GetAllLocations()
	if err != nil {
		return cachedData{}, err
	}

	opts := uniqueLocationOptions(locMap)

	cache = cachedData{
		Artists:   artists,
		Locations: locMap,
		LocOpts:   opts,
		FetchedAt: time.Now(),
	}
	return cache, nil
}

type FilterQuery struct {
	Search string

	CreationMin *int
	CreationMax *int

	MembersMin *int
	MembersMax *int

	FirstAlbumMin *time.Time
	FirstAlbumMax *time.Time

	Locations []string
}

func parseIntPtr(v string) *int {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return nil
	}
	return &n
}

func parseDatePtrYYYYMMDD(v string) *time.Time {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", v)
	if err != nil {
		return nil
	}
	return &t
}

func parseFirstAlbumDDMMYYYY(v string) (*time.Time, bool) {
	t, err := time.Parse("02-01-2006", v)
	if err != nil {
		return nil, false
	}
	return &t, true
}

func parseFilterQuery(r *http.Request) FilterQuery {
	q := FilterQuery{}
	q.Search = strings.TrimSpace(r.URL.Query().Get("q"))

	q.CreationMin = parseIntPtr(r.URL.Query().Get("creation_min"))
	q.CreationMax = parseIntPtr(r.URL.Query().Get("creation_max"))

	q.MembersMin = parseIntPtr(r.URL.Query().Get("members_min"))
	q.MembersMax = parseIntPtr(r.URL.Query().Get("members_max"))

	q.FirstAlbumMin = parseDatePtrYYYYMMDD(r.URL.Query().Get("first_album_min"))
	q.FirstAlbumMax = parseDatePtrYYYYMMDD(r.URL.Query().Get("first_album_max"))

	q.Locations = r.URL.Query()["location"]
	for i := range q.Locations {
		q.Locations[i] = strings.TrimSpace(q.Locations[i])
	}

	return q
}

func normalizeLocSegments(s string) []string {
	s = normalizeLocation(s)
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.ToLower(strings.TrimSpace(p))
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func locationMatches(artistLoc string, selected string) bool {
	al := normalizeLocSegments(artistLoc)
	sl := normalizeLocSegments(selected)
	if len(al) == 0 || len(sl) == 0 {
		return false
	}
	if len(sl) > len(al) {
		return false
	}
	start := len(al) - len(sl)
	for i := range sl {
		if al[start+i] != sl[i] {
			return false
		}
	}
	return true
}

func matchesAnyLocation(artistID int, locMap map[int][]string, selected []string) bool {
	if len(selected) == 0 {
		return true
	}
	artistLocs := locMap[artistID]
	for _, sel := range selected {
		if sel == "" {
			continue
		}
		for _, loc := range artistLocs {
			if locationMatches(loc, sel) {
				return true
			}
		}
	}
	return false
}

func applyFilters(artists []models.Artist, locMap map[int][]string, fq FilterQuery) []models.Artist {
	out := make([]models.Artist, 0, len(artists))
	needle := strings.ToLower(fq.Search)

	for _, a := range artists {
		if needle != "" {
			nameOK := strings.Contains(strings.ToLower(a.Name), needle)
			memOK := false
			if !nameOK {
				for _, m := range a.Members {
					if strings.Contains(strings.ToLower(m), needle) {
						memOK = true
						break
					}
				}
			}
			if !nameOK && !memOK {
				continue
			}
		}

		if fq.CreationMin != nil && a.CreationDate < *fq.CreationMin {
			continue
		}
		if fq.CreationMax != nil && a.CreationDate > *fq.CreationMax {
			continue
		}

		mc := len(a.Members)
		if fq.MembersMin != nil && mc < *fq.MembersMin {
			continue
		}
		if fq.MembersMax != nil && mc > *fq.MembersMax {
			continue
		}

		if fq.FirstAlbumMin != nil || fq.FirstAlbumMax != nil {
			d, ok := parseFirstAlbumDDMMYYYY(a.FirstAlbum)
			if !ok {
				continue
			}
			if fq.FirstAlbumMin != nil && d.Before(*fq.FirstAlbumMin) {
				continue
			}
			if fq.FirstAlbumMax != nil && d.After(*fq.FirstAlbumMax) {
				continue
			}
		}

		if !matchesAnyLocation(a.ID, locMap, fq.Locations) {
			continue
		}

		out = append(out, a)
	}

	return out
}

func uniqueLocationOptions(locMap map[int][]string) []string {
	set := map[string]struct{}{}
	for _, locs := range locMap {
		for _, loc := range locs {
			loc = normalizeLocation(loc)
			parts := strings.Split(loc, ",")
			clean := make([]string, 0, len(parts))
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p != "" {
					clean = append(clean, p)
				}
			}
			for i := 0; i < len(clean); i++ {
				suffix := strings.Join(clean[i:], ", ")
				set[suffix] = struct{}{}
			}
		}
	}

	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

type IndexData struct {
	Artists []models.Artist

	Query string

	CreationMin string
	CreationMax string
	MembersMin  string
	MembersMax  string
	AlbumMin    string
	AlbumMax    string

	LocOptions       []string
	SelectedLocation map[string]bool
}

func Home(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "Erreur template", http.StatusInternalServerError)
		return
	}

	cd, err := getCachedData()
	if err != nil {
		http.Error(w, "Impossible de récupérer les données", http.StatusInternalServerError)
		return
	}

	fq := parseFilterQuery(r)
	res := applyFilters(cd.Artists, cd.Locations, fq)

	sel := map[string]bool{}
	for _, s := range fq.Locations {
		sel[s] = true
	}

	data := IndexData{
		Artists: res,
		Query:   fq.Search,

		CreationMin: r.URL.Query().Get("creation_min"),
		CreationMax: r.URL.Query().Get("creation_max"),
		MembersMin:  r.URL.Query().Get("members_min"),
		MembersMax:  r.URL.Query().Get("members_max"),
		AlbumMin:    r.URL.Query().Get("first_album_min"),
		AlbumMax:    r.URL.Query().Get("first_album_max"),

		LocOptions:       cd.LocOpts,
		SelectedLocation: sel,
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Erreur rendu page", http.StatusInternalServerError)
		return
	}
}

func ArtistDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/artist/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	cd, err := getCachedData()
	if err != nil {
		http.Error(w, "Impossible de récupérer les données", http.StatusInternalServerError)
		return
	}

	var found *models.Artist
	for _, a := range cd.Artists {
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

	var geoLocations []GeoLocation
	for _, loc := range locations {
		gl, err := GeocodeOpenCage(loc)
		if err == nil {
			geoLocations = append(geoLocations, gl)
		} else {
			log.Println("Geocode error:", err)
		}
	}

	geoJSONBytes, err := json.Marshal(geoLocations)
	if err != nil {
		log.Println("GeoJSON error:", err)
		geoJSONBytes = []byte("[]")
	}

	data := struct {
		Artist    *models.Artist
		Dates     []string
		Locations []string
		GeoJSON   template.JS

		Query string
	}{
		Artist:    found,
		Dates:     dates,
		Locations: locations,
		GeoJSON:   template.JS(geoJSONBytes),
		Query:     strings.TrimSpace(r.URL.Query().Get("q")),
	}

	tmpl, err := template.ParseFiles("templates/artist.html")
	if err != nil {
		http.Error(w, "Erreur template", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Erreur rendu page", http.StatusInternalServerError)
		return
	}
}

func Suggest(w http.ResponseWriter, r *http.Request) {
	raw := strings.TrimSpace(r.URL.Query().Get("q"))
	q := strings.ToLower(raw)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if q == "" {
		_ = json.NewEncoder(w).Encode([]any{})
		return
	}

	cd, err := getCachedData()
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

	for _, a := range cd.Artists {
		if strings.HasPrefix(strings.ToLower(a.Name), q) {
			out = append(out, SuggestItem{ID: a.ID, Name: a.Name})
			if len(out) >= limit {
				break
			}
		}
	}

	_ = json.NewEncoder(w).Encode(out)
}
