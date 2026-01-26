package main

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func FavoritesHandler(w http.ResponseWriter, r *http.Request) {
	//  du GET ici pr afficher la page
	if r.Method != http.MethodGet {
		renderError(w, r, http.StatusMethodNotAllowed, "Méthode non autorisée", "Seule la méthode GET est acceptée.")
		return
	}

	// jrecup les ID stocker ds les cookies du gars
	favoriteIDs := getFavoriteIDs(r)
	allArtists, err := FetchArtists()
	if err != nil {
		renderError(w, r, http.StatusInternalServerError, "Erreur serveur", "Impossible de récupérer les artistes.")
		return
	}

	// jboucle sur tt les artistes pr garder ke ceux kil a mis en fav
	var favoriteArtists []Artist
	for _, artist := range allArtists {
		for _, favID := range favoriteIDs {
			if artist.ID == favID {
				favoriteArtists = append(favoriteArtists, artist)
				break
			}
		}
	}

	// jcharge le template html
	tmpl, err := parseTemplate("templates/favorites.html")
	if err != nil {
		renderError(w, r, http.StatusInternalServerError, "Erreur serveur", "Impossible de charger la page.")
		return
	}

	data := PageData{
		Theme: getThemeClass(r),
		Data: map[string]interface{}{
			"Favorites": favoriteArtists,
			"Count":     len(favoriteArtists),
		},
	}
	_ = tmpl.Execute(w, data)
}

func ToggleFavoriteHandler(w http.ResponseWriter, r *http.Request) {
	// la par contre faut du POST pcq on modifie des trucs
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// jrecup lartist_id depuis le formulaire du bouton
	artistIDStr := r.FormValue("artist_id")
	artistID, err := strconv.Atoi(artistIDStr)
	if err != nil {
		http.Error(w, "Invalid artist ID", http.StatusBadRequest)
		return
	}

	// jlis le cookie favorites ki contient un JSON (ex: [1, 7, 12])
	// sil est absent ou peter, la fct renvoie une liste vide
	favoriteIDs := getFavoriteIDs(r)

	// systeme de toggle : si lid est deja la jle vire, sinon jrajoute
	found := false
	var newFavorites []int
	for _, id := range favoriteIDs {
		if id == artistID {
			found = true
			continue // jle met pas ds la nvlle liste donc ca le suppr
		}
		newFavorites = append(newFavorites, id)
	}

	if !found {
		newFavorites = append(newFavorites, artistID)
	}

	// jtransforme la liste en texte pr la foutre ds le cookie
	// json.Marshal convertit mon []int en string genre "[1,7,12]"
	favoriteJSON, _ := json.Marshal(newFavorites)

	// jecris le cookie ds la reponse HTTP pr le navigateur
	http.SetCookie(w, &http.Cookie{
		Name:   "favorites",          // le ptit nom du cookie
		Value:  string(favoriteJSON), // le contenu json en string
		Path:   "/",                  // dispo partout sur le site
		MaxAge: 365 * 24 * 60 * 60,   // il expire ds 1 an
	})

	// jrenvoie lutilisateur sur la page ou il etait juste avant
	referer := r.Header.Get("Referer")
	if referer == "" {
		referer = "/artists"
	}
	http.Redirect(w, r, referer, http.StatusSeeOther)
}

// fct outil pr extraire les ID du cookie proprement
func getFavoriteIDs(r *http.Request) []int {
	cookie, err := r.Cookie("favorites")
	if err != nil {
		return []int{} // si pas de cookie = liste vide
	}

	var favoriteIDs []int
	// jdecode le json du cookie pr en faire une liste de int en Go
	if err := json.Unmarshal([]byte(cookie.Value), &favoriteIDs); err != nil {
		return []int{}
	}

	return favoriteIDs
}

// pr savoir si un artiste est deja ds les fav (pr l'icone coeur)
func isFavorite(r *http.Request, artistID int) bool {
	favoriteIDs := getFavoriteIDs(r)
	for _, id := range favoriteIDs {
		if id == artistID {
			return true
		}
	}
	return false
}

// jprepare une map de status pr le template (vrai/faux pr chaque id)
func GetFavoriteStatus(r *http.Request, artists []Artist) map[int]bool {
	favoriteIDs := getFavoriteIDs(r)
	status := make(map[int]bool)

	for _, artist := range artists {
		status[artist.ID] = false
		for _, favID := range favoriteIDs {
			if artist.ID == favID {
				status[artist.ID] = true
				break
			}
		}
	}

	return status
}
