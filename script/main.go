package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Bienvenue sur mon serveur Go !")
	})

	fmt.Println("Serveur démarré http://127.0.0.1:8080/.")
	http.ListenAndServe(":8080", nil)
}
