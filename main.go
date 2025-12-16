package main

import (
	"fmt"
	"html/template"
	"net/http"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("templates/index.html")
	tmpl.Execute(w, nil)
}

func main() {
	http.HandleFunc("/", indexHandler)
	fmt.Println("Démarrage du serveur sur http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}