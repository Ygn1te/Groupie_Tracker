const searchBar = document.getElementById("searchBar");
const suggestions = document.getElementById("suggestions");
const cardsSelector = ".artistsGrid .card";

// Recherche en temps réel avec le paramètre 'q'
async function updateSuggestions(query) {
  if (query.length < 1) {
    suggestions.classList.remove("active");
    suggestions.innerHTML = "";
    return;
  }
  try {
    const response = await fetch(`/search?q=${encodeURIComponent(query)}`);
    const data = await response.json() || [];
    console.log("Données reçues:", data);
    // Filtre les chanteurs qui COMMENCENT par la requête
    const filtered = data.filter(ch => 
      (ch.name || ch.Name || "").toLowerCase().startsWith(query.toLowerCase())
    );
    console.log("Suggestions filtrées:", filtered);
    
    if (filtered.length > 0) {
      suggestions.innerHTML = filtered
        .slice(0, 10) // Limite à 10 suggestions
        .map(ch => `<li data-id="${ch.id || ch.ID}">${ch.name || ch.Name}</li>`)
        .join("");
      suggestions.classList.add("active");
    } else {
      suggestions.classList.remove("active");
      suggestions.innerHTML = "";
    }
  } catch (err) {
    console.error("Erreur API:", err);
  }
}

function filterArtists() {
  const q = searchBar.value.trim().toLowerCase();
  document.querySelectorAll(cardsSelector).forEach(card => {
    const name = (card.querySelector("h2")?.textContent || "").trim().toLowerCase();
    // Filtre les artistes qui COMMENCENT par la requête
    card.style.display = (!q || name.startsWith(q)) ? "" : "none";
  });
}

async function onInput() {
  const query = searchBar.value.trim();
  await updateSuggestions(query);
  filterArtists();
}

searchBar.addEventListener("input", onInput);
searchBar.addEventListener("keydown", (e) => {
  if (e.key === "Enter") { e.preventDefault(); filterArtists(); }
});
document.addEventListener("DOMContentLoaded", filterArtists);

// Clic sur une suggestion pour la sélectionner
document.addEventListener("click", (e) => {
  if (e.target.tagName === "LI" && e.target.closest("#suggestions")) {
    searchBar.value = e.target.textContent;
    suggestions.classList.remove("active");
    suggestions.innerHTML = "";
    filterArtists();
  }
});

// Fermer les suggestions quand on clique ailleurs
document.addEventListener("click", (e) => {
  if (!e.target.closest("#searchBar") && !e.target.closest("#suggestions")) {
    suggestions.classList.remove("active");
  }
});