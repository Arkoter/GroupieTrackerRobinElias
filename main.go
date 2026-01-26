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

const (
	artistsAPI  = "https://groupietrackers.herokuapp.com/api/artists"
	relationAPI = "https://groupietrackers.herokuapp.com/api/relation"
)

func FetchArtists() []Artist {
	resp, err := http.Get(artistsAPI)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var artists []Artist
	if err := json.NewDecoder(resp.Body).Decode(&artists); err != nil {
		return nil
	}
	return artists
}

func FetchArtistDetail(id int) *ArtistDetail {
	artists := FetchArtists()
	if artists == nil {
		return nil
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
		return nil
	}

	relURL := relationAPI + "/" + strconv.Itoa(id)
	relResp, err := http.Get(relURL)
	if err != nil {
		return &ArtistDetail{Artist: artist, DatesLocations: map[string][]string{}}
	}
	defer relResp.Body.Close()

	var relation Relation
	if err := json.NewDecoder(relResp.Body).Decode(&relation); err != nil {
		return &ArtistDetail{Artist: artist, DatesLocations: map[string][]string{}}
	}

	if relation.DatesLocations == nil {
		relation.DatesLocations = map[string][]string{}
	}

	return &ArtistDetail{
		Artist:         artist,
		DatesLocations: relation.DatesLocations,
	}
}

func getThemeClass(r *http.Request) string {
	cookie, err := r.Cookie("theme")
	if err != nil || cookie.Value == "light" {
		return "light-theme"
	}
	return "dark-theme"
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
		http.Redirect(w, r, "/artists", http.StatusSeeOther)
		return
	}

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "Erreur de template", http.StatusInternalServerError)
		return
	}

	data := PageData{
		Theme: getThemeClass(r),
		Data:  nil,
	}
	_ = tmpl.Execute(w, data)
}

func artistsHandler(w http.ResponseWriter, r *http.Request) {
	artists := FetchArtists()
	if artists == nil {
		http.Error(w, "Impossible de recuperer les artistes.", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/artists.html")
	if err != nil {
		http.Error(w, "Erreur de template", http.StatusInternalServerError)
		return
	}

	data := PageData{
		Theme: getThemeClass(r),
		Data:  artists,
	}
	_ = tmpl.Execute(w, data)
}

func artistDetailHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/artist/")
	id, err := strconv.Atoi(path)
	if err != nil {
		http.Redirect(w, r, "/artists", http.StatusSeeOther)
		return
	}

	artistDetail := FetchArtistDetail(id)
	if artistDetail == nil {
		http.NotFound(w, r)
		return
	}

	tmpl, err := template.ParseFiles("templates/artist_detail.html")
	if err != nil {
		http.Error(w, "Erreur de template", http.StatusInternalServerError)
		return
	}

	data := PageData{
		Theme: getThemeClass(r),
		Data:  artistDetail,
	}
	_ = tmpl.Execute(w, data)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	raw := strings.TrimSpace(r.URL.Query().Get("query"))
	query := strings.ToLower(raw)

	if query == "" {
		http.Redirect(w, r, "/artists", http.StatusSeeOther)
		return
	}

	allArtists := FetchArtists()
	var results []Artist
	if allArtists != nil {
		for _, artist := range allArtists {
			if strings.Contains(strings.ToLower(artist.Name), query) {
				results = append(results, artist)
				continue
			}
			for _, member := range artist.Members {
				if strings.Contains(strings.ToLower(member), query) {
					results = append(results, artist)
					break
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
		http.Error(w, "Erreur de template", http.StatusInternalServerError)
		return
	}

	data := PageData{
		Theme: getThemeClass(r),
		Data:  searchData,
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
	http.HandleFunc("/toggle-theme", toggleThemeHandler)

	println("http://localhost:8080")
	_ = http.ListenAndServe(":8080", nil)
}
