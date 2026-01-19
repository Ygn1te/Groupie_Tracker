package main

import (
	"log"
	"net/http"

	"groupie-tracker/handlers"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Pas de fichier .env trouvé (OK si variables déjà définies).")
	}

	http.HandleFunc("/", handlers.Home)
	http.HandleFunc("/artist/", handlers.ArtistDetail)
	http.HandleFunc("/search", handlers.Search)

	log.Println("Listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
