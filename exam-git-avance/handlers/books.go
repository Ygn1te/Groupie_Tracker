
package handlers

import (
    "html/template"
    "net/http"
    "strconv"
    "strings"

    "exam-git-avance/data"
)

func Home(w http.ResponseWriter, r *http.Request) {
    tmpl := template.Must(template.ParseFiles("templates/layout.html", "templates/index.html"))
    tmpl.Execute(w, nil)
}

func Books(w http.ResponseWriter, r *http.Request) {
    tmpl := template.Must(template.ParseFiles("templates/layout.html", "templates/books.html"))
    tmpl.Execute(w, data.Books)
}

func Book(w http.ResponseWriter, r *http.Request) {
    idStr := strings.TrimPrefix(r.URL.Path, "/book/")
    id, _ := strconv.Atoi(idStr)

    var book data.Book
    for _, b := range data.Books {
        if b.ID == id {
            book = b
            break
        }
    }

    tmpl := template.Must(template.ParseFiles("templates/layout.html", "templates/book.html"))
    tmpl.Execute(w, book)
}
