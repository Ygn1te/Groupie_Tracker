package main

import (
	"groupie-tracker/handlers"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", handlers.Home)
	http.HandleFunc("/artist/", handlers.ArtistDetail)
	http.HandleFunc("/search", handlers.Search)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// optional health http.HandleFunc("/health", handlers.Health)

	log.Println("Listening on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
