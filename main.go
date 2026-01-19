package main

import (
	"encoding/json"
	"fmt"
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

	relURL := fmt.Sprintf("%s/%d", relationAPI, id)
	relResp, err := http.Get(relURL)
	if err != nil {
		return &ArtistDetail{
			Artist:         artist,
			DatesLocations: map[string][]string{},
		}
	}
	defer relResp.Body.Close()

	var relation Relation
	if err := json.NewDecoder(relResp.Body).Decode(&relation); err != nil {
		return &ArtistDetail{
			Artist:         artist,
			DatesLocations: map[string][]string{},
		}
	}

	if relation.DatesLocations == nil {
		relation.DatesLocations = map[string][]string{}
	}

	return &ArtistDetail{
		Artist:         artist,
		DatesLocations: relation.DatesLocations,
	}
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

	_ = tmpl.Execute(w, nil)
}

func artistsHandler(w http.ResponseWriter, r *http.Request) {
	artists := FetchArtists()
	if artists == nil {
		http.Error(w, "Impossible de recuperer les artistes", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/artists.html")
	if err != nil {
		http.Error(w, "Erreur de template", http.StatusInternalServerError)
		return
	}

	_ = tmpl.Execute(w, artists)
}

func artistDetailHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/artist/")
	if path == "" || path == "/" {
		http.Redirect(w, r, "/artists", http.StatusSeeOther)
		return
	}

	id, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	artistDetail := FetchArtistDetail(id)
	if artistDetail == nil {
		http.Error(w, "Artiste non trouve", http.StatusNotFound)
		return
	}

	tmpl, err := template.ParseFiles("templates/artist_detail.html")
	if err != nil {
		http.Error(w, "Erreur de template", http.StatusInternalServerError)
		return
	}

	_ = tmpl.Execute(w, artistDetail)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	raw := strings.TrimSpace(r.URL.Query().Get("query"))
	query := strings.ToLower(raw)

	if query == "" {
		http.Redirect(w, r, "/artists", http.StatusSeeOther)
		return
	}

	allArtists := FetchArtists()
	if allArtists == nil {
		http.Error(w, "Impossible de recuperer les artistes", http.StatusInternalServerError)
		return
	}

	var results []Artist
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

	data := struct {
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

	_ = tmpl.Execute(w, data)
}

func filterHandler(w http.ResponseWriter, r *http.Request) {
	allArtists := FetchArtists()
	if allArtists == nil {
		http.Error(w, "Impossible de recuperer les artistes", http.StatusInternalServerError)
		return
	}

	yearStr := r.URL.Query().Get("year")
	decadeStr := r.URL.Query().Get("decade")
	membersStr := r.URL.Query().Get("members")

	yearActive := yearStr != ""
	decadeActive := decadeStr != ""
	membersActive := membersStr != ""

	if !yearActive && !decadeActive && !membersActive {
		tmpl, err := template.ParseFiles("templates/artists.html")
		if err != nil {
			http.Error(w, "Erreur de template", http.StatusInternalServerError)
			return
		}
		_ = tmpl.Execute(w, allArtists)
		return
	}

	var year int
	var decade int
	var members int
	var membersAtLeastFive bool

	if yearActive {
		v, err := strconv.Atoi(yearStr)
		if err != nil {
			http.Error(w, "Parametre year invalide", http.StatusBadRequest)
			return
		}
		year = v
	}

	if decadeActive {
		v, err := strconv.Atoi(decadeStr)
		if err != nil {
			http.Error(w, "Parametre decade invalide", http.StatusBadRequest)
			return
		}
		decade = v
	}

	if membersActive {
		if membersStr == "5+" {
			membersAtLeastFive = true
		} else {
			v, err := strconv.Atoi(membersStr)
			if err != nil {
				http.Error(w, "Parametre members invalide", http.StatusBadRequest)
				return
			}
			members = v
		}
	}

	var filtered []Artist
	for _, artist := range allArtists {
		if yearActive && artist.CreationDate != year {
			continue
		}
		if decadeActive && !(artist.CreationDate >= decade && artist.CreationDate < decade+10) {
			continue
		}
		if membersActive {
			if membersAtLeastFive {
				if len(artist.Members) < 5 {
					continue
				}
			} else if len(artist.Members) != members {
				continue
			}
		}
		filtered = append(filtered, artist)
	}

	tmpl, err := template.ParseFiles("templates/artists.html")
	if err != nil {
		http.Error(w, "Erreur de template", http.StatusInternalServerError)
		return
	}

	_ = tmpl.Execute(w, filtered)
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/artists", artistsHandler)
	http.HandleFunc("/artist/", artistDetailHandler)
	http.HandleFunc("/search", searchHandler)
	http.HandleFunc("/filter", filterHandler)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	_ = http.ListenAndServe(":8080", nil)
}
