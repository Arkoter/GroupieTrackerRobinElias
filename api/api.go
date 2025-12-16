package api

import (
	"encoding/json"
	"net/http"
)

type Artist struct {
	ID            int      `json:"id"`
	Image         string   `json:"image"`
	Name          string   `json:"name"`
	CreationDate  int      `json:"creationDate"`
	Members       []string `json:"members"`
}

const artistsAPI = "https://groupietrackers.herokuapp.com/api/artists"

func FetchArtists() []Artist {
	resp, _ := http.Get(artistsAPI)
	defer resp.Body.Close()

	var artists []Artist
	json.NewDecoder(resp.Body).Decode(&artists)

	return artists
}