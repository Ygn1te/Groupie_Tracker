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

	addr := ":8080"
	log.Println("Listening on", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
