
package data

type Book struct {
    ID    int
    Title string
    Author string
    Description string
}

var Books = []Book{
    {1, "Le Petit Prince", "Antoine de Saint-Exupéry", "Un aviateur rencontre un petit prince venu d'une autre planète."},
    {2, "1984", "George Orwell", "Un roman dystopique sur une société totalitaire."},
    {3, "L'Étranger", "Albert Camus", "Meursault, un homme indifférent, commet un meurtre absurde."},
}
