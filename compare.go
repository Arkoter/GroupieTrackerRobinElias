package main

import (
	"net/http"
	"strconv"
)

// la structure pr stocker le resultat du match entre 2 artistes
type ComparisonData struct {
	Artist1          *ArtistDetail
	Artist2          *ArtistDetail
	CommonLocations  []string // les villes ou les2 ont jouer
	UniqueLocations1 []string // villes solo artiste 1
	UniqueLocations2 []string // villes solo artiste 2
}

// handler pr afficher la page de selection (les 2 dropdowns)
func CompareHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		renderError(w, r, http.StatusMethodNotAllowed, "Methode pas autorisee", "Faut du GET")
		return
	}

	// jrecup la liste complete pr que lutilisateur puisse choisir
	artists, err := FetchArtists()
	if err != nil {
		renderError(w, r, http.StatusInternalServerError, "Err serv", "Jai pas pu recup les artistes.")
		return
	}

	tmpl, err := parseTemplate("templates/compare.html")
	if err != nil {
		renderError(w, r, http.StatusInternalServerError, "Err serv", "Template introuvable.")
		return
	}

	data := PageData{
		Theme: getThemeClass(r),
		Data:  artists,
	}
	_ = tmpl.Execute(w, data)
}

// handler pr afficher le resultat du duel
func CompareResultHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		renderError(w, r, http.StatusMethodNotAllowed, "Methode pas autorisee", "Faut du GET")
		return
	}

	// jrecup artist1 et artist2 depuis la query string (l'url quoi)
	id1Str := r.URL.Query().Get("artist1")
	id2Str := r.URL.Query().Get("artist2")

	if id1Str == "" || id2Str == "" {
		renderError(w, r, http.StatusBadRequest, "Param manquants", "Choisis 2 artistes stp.")
		return
	}

	// je convertis en int + check si cest des chiffres valides
	id1, err1 := strconv.Atoi(id1Str)
	id2, err2 := strconv.Atoi(id2Str)

	if err1 != nil || err2 != nil {
		renderError(w, r, http.StatusBadRequest, "Param invalides", "Les ID doivent etre des nombres.")
		return
	}

	// je verifie que cest pas le meme artiste (aucun interet sinon)
	if id1 == id2 {
		renderError(w, r, http.StatusBadRequest, "Err comparaison", "Selectionne 2 mecs differents.")
		return
	}

	// jrecup les infos via FetchArtistDetail(id) pr avoir les concerts (DatesLocations)
	artist1, err := FetchArtistDetail(id1)
	if err != nil || artist1 == nil {
		renderError(w, r, http.StatusNotFound, "Pas trouver", "Lartiste 1 existe pas.")
		return
	}

	artist2, err := FetchArtistDetail(id2)
	if err != nil || artist2 == nil {
		renderError(w, r, http.StatusNotFound, "Pas trouver", "Lartiste 2 existe pas.")
		return
	}

	// jcalcule les differences et les points communs ici
	comparison := compareArtists(artist1, artist2)

	tmpl, err := parseTemplate("templates/compare_result.html")
	if err != nil {
		renderError(w, r, http.StatusInternalServerError, "Err serv", "Template mort.")
		return
	}

	// jepasse les 2 artistes + le resultat au template
	data := PageData{
		Theme: getThemeClass(r),
		Data:  comparison,
	}
	_ = tmpl.Execute(w, data)
}

// la fct qui fait tout le boulot de tri pr les lieux de concerts
func compareArtists(artist1, artist2 *ArtistDetail) ComparisonData {
	locations1 := make(map[string]bool)
	locations2 := make(map[string]bool)

	// jefous tout ds des maps pr que ce soit + simple a comparer
	for loc := range artist1.DatesLocations {
		locations1[loc] = true
	}

	for loc := range artist2.DatesLocations {
		locations2[loc] = true
	}

	var common []string
	var unique1 []string
	var unique2 []string

	// jeboucle sur le 1er pr voir ce qui est commun ou solo
	for loc := range locations1 {
		if locations2[loc] {
			common = append(common, loc)
		} else {
			unique1 = append(unique1, loc)
		}
	}

	// jeboucle sur le 2eme pr chopper ses villes uniques a lui
	for loc := range locations2 {
		if !locations1[loc] {
			unique2 = append(unique2, loc)
		}
	}

	return ComparisonData{
		Artist1:          artist1,
		Artist2:          artist2,
		CommonLocations:  common,
		UniqueLocations1: unique1,
		UniqueLocations2: unique2,
	}
}
