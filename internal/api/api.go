package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type Artist struct {
	ID           int      `json:"id"`
	Image        string   `json:"image"`
	Name         string   `json:"name"`
	CreationDate int      `json:"creationDate"`
	FirstAlbum   string   `json:"firstAlbum"`
	Members      []string `json:"members"`
}

func main() {
	fmt.Println("Test de connexion a l'API Groupie Tracker")
	fmt.Println("========================================")
	fmt.Println()

	url := "https://groupietrackers.herokuapp.com/api/artists"

	resp, err := http.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erreur: sa focntionne pa\nDetails: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	fmt.Printf("Connexion OK (Status: %s)\n", resp.Status)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Fprintf(os.Stderr, "Erreur: sa focntionne pa(%d)\n", resp.StatusCode)
		os.Exit(1)
	}

	var artists []Artist
	if err := json.NewDecoder(resp.Body).Decode(&artists); err != nil {
		fmt.Fprintf(os.Stderr, "Erreur: sa focntionne pa\nDetails: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("Nombre dartistes: %d\n", len(artists))

	if len(artists) == 0 {
		return
	}

	fmt.Println()
	fmt.Println("Premier artistes:")
	limit := 5
	if len(artists) < limit {
		limit = len(artists)
	}

	for i := 0; i < limit; i++ {
		fmt.Printf("%d. %s (%d) - %d membres\n",
			artists[i].ID,
			artists[i].Name,
			artists[i].CreationDate,
			len(artists[i].Members),
		)
	}
}
