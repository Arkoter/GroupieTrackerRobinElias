package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"groupie-tracker/api"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("templates/index.html")
	tmpl.Execute(w, nil)
}

// (Partie 2)
func artistsHandler(w http.ResponseWriter, r *http.Request) {
	artists := api.FetchArtists()
	tmpl, err := template.ParseFiles("templates/artists.html")
	if err != nil {
		http.Error(w, "Erreur de template", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, artists)
}

// (Partie 3)
func artistDetailHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/artist/")

	if path == "" || path == "/" {
		http.Redirect(w, r, "/artists", http.StatusSeeOther)
		return
	}

	id, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "ID invalide. Utilisez /artist/1, /artist/2, etc.", http.StatusBadRequest)
		return
	}

	// Récup les donnais de lartiste
	artistDetail := api.FetchArtistDetail(id)
	if artistDetail == nil {
		http.Error(w, "Artiste non trouvé", http.StatusNotFound)
		return
	}

	tmpl, err := template.ParseFiles("templates/artist_detail.html")
	if err != nil {
		http.Error(w, "Erreur de template", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, artistDetail)
}

// (Partie 4)
func searchHandler(w http.ResponseWriter, r *http.Request) {
	//je recup le paramatre avec r.URL.Query().Get("query") et je le passe en minuscule en faisan strings.ToLower
	query := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("query")))

	if query == "" {
		http.Redirect(w, r, "/artists", http.StatusSeeOther)
		return
	}

	// Récup tous les artiste
	allArtists := api.FetchArtists()

	// Filtrer les artistes selon la recherche
	var results []api.Artist

	//je parcours tout les artiste
	for _, artist := range allArtists {
		// Recherche par nom d'artiste
		//je verif si le nom de lartiste contient la recherche en utilisant string.cotnains
		if strings.Contains(strings.ToLower(artist.Name), query) {
			results = append(results, artist)
			continue
		}

		// Recherche par membre
		for _, member := range artist.Members {
			//je verif si un des membres contient la recherche en utilisant string.cotnains
			if strings.Contains(strings.ToLower(member), query) {
				results = append(results, artist)
				break
			}
		}
	}

	// Creer les donnais pour le template
	data := struct {
		Query   string
		Results []api.Artist
		Count   int
	}{
		Query:   r.URL.Query().Get("query"),
		Results: results,
		Count:   len(results),
	}

	tmpl, err := template.ParseFiles("templates/search.html")
	if err != nil {
		http.Error(w, "Erreur de template", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, data)
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/artists", artistsHandler)
	http.HandleFunc("/artist/", artistDetailHandler)
	http.HandleFunc("/search", searchHandler)

	fmt.Println("Démarrage du serveur sur http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
