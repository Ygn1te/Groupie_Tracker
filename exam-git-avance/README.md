# Mini bibliothèque – Examen Git Avancé

Projet Go pour un mini serveur web affichant une liste de livres.

## Structure
- `main.go` : point d'entrée du serveur
- `data/books.go` : données des livres
- `handlers/books.go` : routes HTTP
- `templates/` : templates HTML
- `static/style.css` : styles

## Lancer le projet
```
go mod init main
go mod tidy
go run main.go
```

## Note
Le fichier `go.mod` n'est volontairement **pas inclus**, conformément aux consignes du TP.
