package main

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

type Artist struct {
	ID           int      `json:"id"`
	Image        string   `json:"image"`
	Name         string   `json:"name"`
	CreationDate int      `json:"creationDate"`
	FirstAlbum   string   `json:"firstAlbum"`
	Members      []string `json:"members"`
}

type Relation struct {
	DatesLocations map[string][]string `json:"datesLocations"`
}

type ArtistDetail struct {
	Artist
	DatesLocations map[string][]string
}

type PageData struct {
	Theme string
	Data  interface{}
}

type ErrorPageData struct {
	Theme   string
	Code    int
	Message string
	Details string
}

const (
	artistsAPI  = "https://groupietrackers.herokuapp.com/api/artists"
	relationAPI = "https://groupietrackers.herokuapp.com/api/relation"
)

var cachedArtists []Artist
var cachedRelations = make(map[int]map[string][]string)

func FetchArtists() ([]Artist, error) {
	if cachedArtists != nil {
		return cachedArtists, nil
	}

	resp, err := http.Get(artistsAPI)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, err
	}

	var artists []Artist
	if err := json.NewDecoder(resp.Body).Decode(&artists); err != nil {
		return nil, err
	}

	cachedArtists = artists
	return artists, nil
}

func FetchArtistDetail(id int) (*ArtistDetail, error) {
	artists, err := FetchArtists()
	if err != nil {
		return nil, err
	}

	var artist Artist
	found := false
	for _, a := range artists {
		if a.ID == id {
			artist = a
			found = true
			break
		}
	}
	if !found {
		return nil, nil
	}

	if relations, ok := cachedRelations[id]; ok {
		return &ArtistDetail{
			Artist:         artist,
			DatesLocations: relations,
		}, nil
	}

	relURL := relationAPI + "/" + strconv.Itoa(id)
	relResp, err := http.Get(relURL)
	if err != nil {
		return &ArtistDetail{Artist: artist, DatesLocations: map[string][]string{}}, nil
	}
	defer relResp.Body.Close()

	var relation Relation
	if err := json.NewDecoder(relResp.Body).Decode(&relation); err != nil {
		return &ArtistDetail{Artist: artist, DatesLocations: map[string][]string{}}, nil
	}

	if relation.DatesLocations == nil {
		relation.DatesLocations = map[string][]string{}
	}

	cachedRelations[id] = relation.DatesLocations

	return &ArtistDetail{
		Artist:         artist,
		DatesLocations: relation.DatesLocations,
	}, nil
}

func getThemeClass(r *http.Request) string {
	cookie, err := r.Cookie("theme")
	if err != nil || cookie.Value == "light" {
		return "light-theme"
	}
	return "dark-theme"
}

func renderError(w http.ResponseWriter, r *http.Request, code int, message, details string) {
	tmpl, err := template.ParseFiles("templates/error.html")
	if err != nil {
		http.Error(w, message, code)
		return
	}

	w.WriteHeader(code)
	data := ErrorPageData{
		Theme:   getThemeClass(r),
		Code:    code,
		Message: message,
		Details: details,
	}
	_ = tmpl.Execute(w, data)
}

func toggleThemeHandler(w http.ResponseWriter, r *http.Request) {
	currentTheme := "light"
	cookie, err := r.Cookie("theme")
	if err == nil {
		currentTheme = cookie.Value
	}

	newTheme := "light"
	if currentTheme == "light" {
		newTheme = "dark"
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "theme",
		Value: newTheme,
		Path:  "/",
	})

	referer := r.Header.Get("Referer")
	if referer == "" {
		referer = "/"
	}
	http.Redirect(w, r, referer, http.StatusSeeOther)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		renderError(w, r, http.StatusNotFound, "Page non trouvée", "La page que vous recherchez n'existe pas.")
		return
	}

	if r.Method != http.MethodGet {
		renderError(w, r, http.StatusMethodNotAllowed, "Méthode non autorisée", "Seule la méthode GET est acceptée.")
		return
	}

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		renderError(w, r, http.StatusInternalServerError, "Erreur serveur", "Impossible de charger la page.")
		return
	}

	data := PageData{
		Theme: getThemeClass(r),
		Data:  nil,
	}
	_ = tmpl.Execute(w, data)
}

func artistsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		renderError(w, r, http.StatusMethodNotAllowed, "Méthode non autorisée", "Seule la méthode GET est acceptée.")
		return
	}

	artists, err := FetchArtists()
	if err != nil {
		renderError(w, r, http.StatusInternalServerError, "Erreur serveur", "Impossible de récupérer les artistes depuis l'API.")
		return
	}

	// Filtres
	minYear := r.URL.Query().Get("min_year")
	maxYear := r.URL.Query().Get("max_year")
	members := r.URL.Query().Get("members")
	location := r.URL.Query().Get("location")

	filtered := filterArtists(artists, minYear, maxYear, members, location)

	tmpl, err := template.ParseFiles("templates/artists.html")
	if err != nil {
		renderError(w, r, http.StatusInternalServerError, "Erreur serveur", "Impossible de charger la page.")
		return
	}

	data := PageData{
		Theme: getThemeClass(r),
		Data: map[string]interface{}{
			"Artists":  filtered,
			"MinYear":  minYear,
			"MaxYear":  maxYear,
			"Members":  members,
			"Location": location,
		},
	}
	_ = tmpl.Execute(w, data)
}

func filterArtists(artists []Artist, minYear, maxYear, members, location string) []Artist {
	var filtered []Artist

	for _, artist := range artists {
		// Filtre annee min
		if minYear != "" {
			min, err := strconv.Atoi(minYear)
			if err == nil && artist.CreationDate < min {
				continue
			}
		}

		// Filtre annee max
		if maxYear != "" {
			max, err := strconv.Atoi(maxYear)
			if err == nil && artist.CreationDate > max {
				continue
			}
		}

		// Filtre nombre demembres
		if members != "" {
			memberCount, err := strconv.Atoi(members)
			if err == nil && len(artist.Members) != memberCount {
				continue
			}
		}

		// Filtre par lieu (faut vrifier les relations)
		if location != "" {
			locationLower := strings.ToLower(location)
			hasLocation := false

			detail, _ := FetchArtistDetail(artist.ID)
			if detail != nil {
				for loc := range detail.DatesLocations {
					if strings.Contains(strings.ToLower(loc), locationLower) {
						hasLocation = true
						break
					}
				}
			}

			if !hasLocation {
				continue
			}
		}

		filtered = append(filtered, artist)
	}

	return filtered
}

func artistDetailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		renderError(w, r, http.StatusMethodNotAllowed, "Méthode non autorisée", "Seule la méthode GET est acceptée.")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/artist/")
	id, err := strconv.Atoi(path)
	if err != nil {
		renderError(w, r, http.StatusBadRequest, "Paramètre invalide", "L'ID de l'artiste doit être un nombre.")
		return
	}

	artistDetail, err := FetchArtistDetail(id)
	if err != nil {
		renderError(w, r, http.StatusInternalServerError, "Erreur serveur", "Impossible de récupérer les détails de l'artiste.")
		return
	}

	if artistDetail == nil {
		renderError(w, r, http.StatusNotFound, "Artiste non trouvé", "Aucun artiste ne correspond à cet ID.")
		return
	}

	tmpl, err := template.ParseFiles("templates/artist_detail.html")
	if err != nil {
		renderError(w, r, http.StatusInternalServerError, "Erreur serveur", "Impossible de charger la page.")
		return
	}

	data := PageData{
		Theme: getThemeClass(r),
		Data:  artistDetail,
	}
	_ = tmpl.Execute(w, data)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		renderError(w, r, http.StatusMethodNotAllowed, "Méthode non autorisée", "Seule la méthode GET est acceptée.")
		return
	}

	raw := strings.TrimSpace(r.URL.Query().Get("query"))
	query := strings.ToLower(raw)

	if query == "" {
		http.Redirect(w, r, "/artists", http.StatusSeeOther)
		return
	}

	allArtists, err := FetchArtists()
	if err != nil {
		renderError(w, r, http.StatusInternalServerError, "Erreur serveur", "Impossible de récupérer les artistes.")
		return
	}

	var results []Artist
	for _, artist := range allArtists {
		// Recherche par nom
		if strings.Contains(strings.ToLower(artist.Name), query) {
			results = append(results, artist)
			continue
		}

		// Recherche par membre
		foundMember := false
		for _, member := range artist.Members {
			if strings.Contains(strings.ToLower(member), query) {
				results = append(results, artist)
				foundMember = true
				break
			}
		}
		if foundMember {
			continue
		}

		// Recherche par lieu ou date
		detail, _ := FetchArtistDetail(artist.ID)
		if detail != nil {
			for location, dates := range detail.DatesLocations {
				if strings.Contains(strings.ToLower(location), query) {
					results = append(results, artist)
					break
				}
				for _, date := range dates {
					if strings.Contains(strings.ToLower(date), query) {
						results = append(results, artist)
						break
					}
				}
			}
		}
	}

	searchData := struct {
		Query   string
		Results []Artist
		Count   int
	}{
		Query:   raw,
		Results: results,
		Count:   len(results),
	}

	tmpl, err := template.ParseFiles("templates/search.html")
	if err != nil {
		renderError(w, r, http.StatusInternalServerError, "Erreur serveur", "Impossible de charger la page.")
		return
	}

	data := PageData{
		Theme: getThemeClass(r),
		Data:  searchData,
	}
	_ = tmpl.Execute(w, data)
}

// locationHandler - Événement interactif : clic sur un lieu
func locationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		renderError(w, r, http.StatusMethodNotAllowed, "Méthode non autorisée", "Seule la méthode GET est acceptée.")
		return
	}

	locationQuery := r.URL.Query().Get("loc")
	if locationQuery == "" {
		renderError(w, r, http.StatusBadRequest, "Paramètre manquant", "Le lieu doit être spécifié.")
		return
	}

	allArtists, err := FetchArtists()
	if err != nil {
		renderError(w, r, http.StatusInternalServerError, "Erreur serveur", "Impossible de récupérer les artistes.")
		return
	}

	type ConcertInfo struct {
		Artist Artist
		Dates  []string
	}

	var concerts []ConcertInfo
	locationLower := strings.ToLower(locationQuery)

	for _, artist := range allArtists {
		detail, _ := FetchArtistDetail(artist.ID)
		if detail != nil {
			for location, dates := range detail.DatesLocations {
				if strings.ToLower(location) == locationLower {
					concerts = append(concerts, ConcertInfo{
						Artist: artist,
						Dates:  dates,
					})
					break
				}
			}
		}
	}

	tmpl, err := template.ParseFiles("templates/location.html")
	if err != nil {
		renderError(w, r, http.StatusInternalServerError, "Erreur serveur", "Impossible de charger la page.")
		return
	}

	data := PageData{
		Theme: getThemeClass(r),
		Data: map[string]interface{}{
			"Location": locationQuery,
			"Concerts": concerts,
			"Count":    len(concerts),
		},
	}
	_ = tmpl.Execute(w, data)
}

func main() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/artists", artistsHandler)
	http.HandleFunc("/artist/", artistDetailHandler)
	http.HandleFunc("/search", searchHandler)
	http.HandleFunc("/location", locationHandler)
	http.HandleFunc("/toggle-theme", toggleThemeHandler)

	println("  http://localhost:8080")
	_ = http.ListenAndServe(":8080", nil)
}
