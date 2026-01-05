package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Artist struct {
	ID           int      `json:"id"`
	Image        string   `json:"image"`
	Name         string   `json:"name"`
	CreationDate int      `json:"creationDate"`
	Members      []string `json:"members"`
}

type Locations struct {
	Locations []string `json:"locations"`
}

type Dates struct {
	Dates []string `json:"dates"`
}

type Relation struct {
	DatesLocations map[string][]string `json:"datesLocations"`
}

type ArtistDetail struct {
	Artist
	FirstAlbum     string
	Locations      []string
	Dates          []string
	DatesLocations map[string][]string
}

const artistsAPI = "https://groupietrackers.herokuapp.com/api/artists"
const locationsAPI = "https://groupietrackers.herokuapp.com/api/locations"
const datesAPI = "https://groupietrackers.herokuapp.com/api/dates"
const relationAPI = "https://groupietrackers.herokuapp.com/api/relation"

func FetchArtists() []Artist {
	resp, err := http.Get(artistsAPI)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var artists []Artist
	json.NewDecoder(resp.Body).Decode(&artists)
	return artists
}

func FetchArtistDetail(id int) *ArtistDetail {
	// Récupairer lartist de base
	artists := FetchArtists()
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

	// Récup les locations
	locResp, err := http.Get(fmt.Sprintf("%s/%d", locationsAPI, id))
	if err != nil {
		return nil
	}
	defer locResp.Body.Close()

	var locations Locations
	json.NewDecoder(locResp.Body).Decode(&locations)

	// Récup les date
	dateResp, err := http.Get(fmt.Sprintf("%s/%d", datesAPI, id))
	if err != nil {
		return nil
	}
	defer dateResp.Body.Close()

	var dates Dates
	json.NewDecoder(dateResp.Body).Decode(&dates)

	// Récup les relations
	relResp, err := http.Get(fmt.Sprintf("%s/%d", relationAPI, id))
	if err != nil {
		return nil
	}
	defer relResp.Body.Close()

	var relation Relation
	json.NewDecoder(relResp.Body).Decode(&relation)

	return &ArtistDetail{
		Artist:         artist,
		FirstAlbum:     "",
		Locations:      locations.Locations,
		Dates:          dates.Dates,
		DatesLocations: relation.DatesLocations,
	}
}
