package main

import (
	"groupie_tracker/handlers"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", handlers.Home)
	http.HandleFunc("/artist/", handlers.ArtistDetail)
	http.HandleFunc("/search", handlers.SearchArtists)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("Listening on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
