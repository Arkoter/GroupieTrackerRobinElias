package main

import (
	"fmt"
	"html/template"
	"net/http"

	"groupie-tracker/api"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	artists := api.FetchArtists()

	tmpl, _ := template.ParseFiles("templates/index.html")
	tmpl.Execute(w, artists)
}

func artistsHandler(w http.ResponseWriter, r *http.Request) {
	artists := api.FetchArtists()

	tmpl, _ := template.ParseFiles("templates/artists.html")
	tmpl.Execute(w, artists)
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/artists", artistsHandler)

	fmt.Println("Démarrage du serveur sur http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}