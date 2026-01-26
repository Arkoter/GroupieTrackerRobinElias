package main

// Structures principales
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

// Constantes API
const (
	artistsAPI  = "https://groupietrackers.herokuapp.com/api/artists"
	relationAPI = "https://groupietrackers.herokuapp.com/api/relation"
)
