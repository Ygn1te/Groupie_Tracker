
package main

import (
    "log"
    "net/http"
    "exam-git-avance/handlers"
)

func main() {
    mux := http.NewServeMux()

    mux.HandleFunc("/", handlers.Home)
    mux.HandleFunc("/books", handlers.Books)
    mux.HandleFunc("/book/", handlers.Book)

    log.Println("Server running on http://localhost:8080")
    http.ListenAndServe(":8080", mux)
}
