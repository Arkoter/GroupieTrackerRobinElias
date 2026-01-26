package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// fetchArtistsFromAPI fait la vraie requete a l'API
func fetchArtistsFromAPI() ([]Artist, error) {
	resp, err := http.Get(artistsAPI)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, err
	}

	var artists []Artist
	// jdecode le flux json direct ds la variable
	if err := json.NewDecoder(resp.Body).Decode(&artists); err != nil {
		return nil, err
	}
	return artists, nil
}

// fetchRelationFromAPI fait la vraie requete pr les dates et lieux
func fetchRelationFromAPI(id int) (map[string][]string, error) {
	relURL := relationAPI + "/" + strconv.Itoa(id)
	relResp, err := http.Get(relURL)
	if err != nil {
		return map[string][]string{}, err
	}
	defer relResp.Body.Close()

	var relation Relation
	if err := json.NewDecoder(relResp.Body).Decode(&relation); err != nil {
		return map[string][]string{}, err
	}
	// si cest vide jrenvoie un truc propre pr pas que ca plante
	if relation.DatesLocations == nil {
		relation.DatesLocations = map[string][]string{}
	}
	return relation.DatesLocations, nil
}

// FetchArtists passe par mon cache pr pas saturer l'API
func FetchArtists() ([]Artist, error) {
	return apiCache.GetArtists()
}

// FetchArtistDetail regroupe les infos de base + les relations
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

	relations, err := apiCache.GetRelation(id)
	if err != nil {
		return &ArtistDetail{Artist: artist, DatesLocations: map[string][]string{}}, nil
	}

	return &ArtistDetail{
		Artist:         artist,
		DatesLocations: relations,
	}, nil
}

// jrecup le theme via le cookie pr savoir si jmet la classe dark ou light
func getThemeClass(r *http.Request) string {
	cookie, err := r.Cookie("theme")
	if err != nil || cookie.Value == "light" {
		return "light-theme"
	}
	return "dark-theme"
}

// ma fct pr eviter les crashs et envoyer des pages derreurs propres
func renderError(w http.ResponseWriter, r *http.Request, code int, message, details string) {
	// jessaye de charger mon template special erreur
	tmpl, err := parseTemplate("templates/error.html")
	if err != nil {
		// si mm le template derreur marche pas, jrenvoie un truc brut pr pas crash
		http.Error(w, message, code)
		return
	}

	// jprecise le code HTTP (404, 500, etc) ds le header de la reponse
	w.WriteHeader(code)

	// jfous ttes les infos pr que luser comprenne ce quil a peter
	data := ErrorPageData{
		Theme:   getThemeClass(r),
		Code:    code,
		Message: message,
		Details: details, // ex: 'ID invalide' ou 'Page inexistante'
	}
	_ = tmpl.Execute(w, data)
}

// ex pr la 404 qd la route est pas bonne
func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		// jappel ma fct pr pas que le serv s'arrete
		renderError(w, r, http.StatusNotFound, "Page non trouvée", "La page que vous recherchez n'existe pas.")
		return
	}
	tmpl, _ := parseTemplate("templates/index.html")
	data := PageData{Theme: getThemeClass(r), Data: nil}
	_ = tmpl.Execute(w, data)
}

// 6) EVENEMENT INTERACTIF : qd on switch le theme ca fait une requete serveur
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

	// jsave le choix ds un cookie pr que ca reste au refresh
	http.SetCookie(w, &http.Cookie{
		Name:  "theme",
		Value: newTheme,
		Path:  "/",
	})

	referer := r.Header.Get("Referer") // jrenvoie le gars d'ou il vient
	if referer == "" {
		referer = "/"
	}
	http.Redirect(w, r, referer, http.StatusSeeOther)
}

// 6) EVENEMENT INTERACTIF : filtrer les artistes = requete serveur GET
func artistsHandler(w http.ResponseWriter, r *http.Request) {
	artists, err := FetchArtists()
	if err != nil {
		renderError(w, r, http.StatusInternalServerError, "Err serv", "API down ?")
		return
	}

	// jrecup les params ds l'url (ex: ?min_year=2000)
	minYear := r.URL.Query().Get("min_year")
	maxYear := r.URL.Query().Get("max_year")
	members := r.URL.Query().Get("members")
	location := r.URL.Query().Get("location")

	// le filtrage se fait cote serveur ds cette fct
	filtered := filterArtists(artists, minYear, maxYear, members, location)
	favoriteStatus := GetFavoriteStatus(r, filtered)

	tmpl, _ := parseTemplate("templates/artists.html")
	data := PageData{
		Theme: getThemeClass(r),
		Data: map[string]interface{}{
			"Artists":        filtered,
			"MinYear":        minYear,
			"MaxYear":        maxYear,
			"Members":        members,
			"Location":       location,
			"FavoriteStatus": favoriteStatus,
		},
	}
	_ = tmpl.Execute(w, data)
}

// la fct qui gere tt le tri selon ce que luser a demander
func filterArtists(artists []Artist, minYear, maxYear, members, location string) []Artist {
	var filtered []Artist

	for _, artist := range artists {
		// --- FILTRE PAR INTERVALLE (ANNEE MIN/MAX) ---
		// jconvertis le texte du form en chiffre pr comparer avec CreationDate
		if minYear != "" {
			min, err := strconv.Atoi(minYear)
			if err == nil && artist.CreationDate < min {
				continue // si trop vieux -> jpasse au suivant
			}
		}
		if maxYear != "" {
			max, err := strconv.Atoi(maxYear)
			if err == nil && artist.CreationDate > max {
				continue // si trop recent -> jpasse au suivant
			}
		}

		// --- FILTRE SELECTION MULTIPLE (MEMBRES) ---
		// on check la taille de l'array members pr voir si ca match
		if members != "" {
			memberCount, err := strconv.Atoi(members)
			if err == nil && len(artist.Members) != memberCount {
				continue
			}
		}

		// --- COMBINAISON DES FILTRES ---
		// le truc  c que si un seul filtre echoue, le 'continue' zappe lartiste
		// du coup ca cumule les filtres automatiquement sans forcer
		if location != "" {
			locationLower := strings.ToLower(location)
			hasLocation := false
			// jsuis obliger daller chercher les details pr checker les villes de concert
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

		// si il a survecu a tt les 'continue', jle garde
		filtered = append(filtered, artist)
	}

	return filtered
}

// 6) EVENEMENT INTERACTIF : cliquer "voir details" -> requete /artist/{id}
func artistDetailHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/artist/")
	id, err := strconv.Atoi(path)
	if err != nil {
		renderError(w, r, http.StatusBadRequest, "Err", "ID n'est pas un nombre")
		return
	}

	artistDetail, _ := FetchArtistDetail(id)
	if artistDetail == nil {
		renderError(w, r, http.StatusNotFound, "Pas la", "Artiste inconnu")
		return
	}

	tmpl, _ := parseTemplate("templates/artist_detail.html")
	_ = tmpl.Execute(w, PageData{Theme: getThemeClass(r), Data: artistDetail})
}

// systeme de recherche globale (nom, membres, lieux, dates)
func searchHandler(w http.ResponseWriter, r *http.Request) {
	raw := strings.TrimSpace(r.URL.Query().Get("query"))
	query := strings.ToLower(raw)
	if query == "" {
		http.Redirect(w, r, "/artists", http.StatusSeeOther)
		return
	}

	allArtists, _ := FetchArtists()
	var results []Artist
	for _, artist := range allArtists {
		// si le nom match jle met direct
		if strings.Contains(strings.ToLower(artist.Name), query) {
			results = append(results, artist)
			continue
		}
		// jboucle sur les membres pr voir si un nom match
		foundMember := false
		for _, m := range artist.Members {
			if strings.Contains(strings.ToLower(m), query) {
				results = append(results, artist)
				foundMember = true
				break
			}
		}
		if foundMember {
			continue
		}

		// check ds les concerts si la ville ou date match
		detail, _ := FetchArtistDetail(artist.ID)
		if detail != nil {
			for loc, dates := range detail.DatesLocations {
				if strings.Contains(strings.ToLower(loc), query) {
					results = append(results, artist)
					break
				}
				for _, d := range dates {
					if strings.Contains(strings.ToLower(d), query) {
						results = append(results, artist)
						break
					}
				}
			}
		}
	}
	tmpl, _ := parseTemplate("templates/search.html")
	_ = tmpl.Execute(w, PageData{Theme: getThemeClass(r), Data: map[string]interface{}{"Query": raw, "Results": results, "Count": len(results)}})
}

// 6) EVENEMENT INTERACTIF : clic sur un lieu -> nouvelle page /location?loc=...
func locationHandler(w http.ResponseWriter, r *http.Request) {
	locationQuery := r.URL.Query().Get("loc")
	allArtists, _ := FetchArtists()

	type ConcertInfo struct {
		Artist Artist
		Dates  []string
	}
	var concerts []ConcertInfo
	locationLower := strings.ToLower(locationQuery)

	// jrecherche tt les concerts de tt les artistes pr ce lieu precis
	for _, artist := range allArtists {
		detail, _ := FetchArtistDetail(artist.ID)
		if detail != nil {
			for loc, dates := range detail.DatesLocations {
				if strings.ToLower(loc) == locationLower {
					concerts = append(concerts, ConcertInfo{Artist: artist, Dates: dates})
					break
				}
			}
		}
	}
	tmpl, _ := parseTemplate("templates/location.html")
	_ = tmpl.Execute(w, PageData{Theme: getThemeClass(r), Data: map[string]interface{}{"Location": locationQuery, "Concerts": concerts, "Count": len(concerts)}})
}

func main() {
	// sert les fichiers css et js
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// tt les routes du site
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/artists", artistsHandler)
	http.HandleFunc("/artist/", artistDetailHandler)
	http.HandleFunc("/search", searchHandler)
	http.HandleFunc("/location", locationHandler)
	http.HandleFunc("/toggle-theme", toggleThemeHandler)
	http.HandleFunc("/favorites", FavoritesHandler)
	http.HandleFunc("/toggle-favorite", ToggleFavoriteHandler)
	http.HandleFunc("/compare", CompareHandler)
	http.HandleFunc("/compare-result", CompareResultHandler)
	http.HandleFunc("/cache-info", CacheInfoHandler)
	http.HandleFunc("/refresh-cache", RefreshCacheHandler)

	println("Serveur demarrer sur http://localhost:8080")
	_ = http.ListenAndServe(":8080", nil)
}
