// Suggestions en temps réel
const searchBar = document.getElementById('searchBar');
const suggestionsList = document.getElementById('suggestions');

async function fetchSuggestions(query) {
    if (query.length === 0) {
        suggestionsList.classList.remove('active');
        return;
    }
    
    try {
        const response = await fetch(`/search?q=${encodeURIComponent(query)}`);
        const artists = await response.json();
        displaySuggestions(artists);
    } catch (error) {
        console.error('Erreur:', error);
        suggestionsList.classList.remove('active');
    }
}

function displaySuggestions(artists) {
    suggestionsList.innerHTML = '';
    
    if (!artists || artists.length === 0) {
        suggestionsList.classList.remove('active');
        return;
    }
    
    artists.forEach(artist => {
        const li = document.createElement('li');
        li.textContent = artist.name;
        
        li.addEventListener('click', () => {
            window.location.href = `/?q=${encodeURIComponent(artist.name)}`;
        });
        
        suggestionsList.appendChild(li);
    });
    
    suggestionsList.classList.add('active');
}

searchBar.addEventListener('input', (e) => {
    const query = e.target.value.trim();
    fetchSuggestions(query);
});

// Fermer les suggestions au clic ailleurs
document.addEventListener('click', (e) => {
    if (e.target !== searchBar) {
        suggestionsList.classList.remove('active');
    }
});

const cardsSelector = ".artistsGrid .card";

document.querySelectorAll('.card').forEach(card => { // Effet 3D au survol
  card.addEventListener('mousemove', e => {
    const rect = card.getBoundingClientRect();
    const x = e.clientX - rect.left; // position X dans la carte
    const y = e.clientY - rect.top;  // position Y dans la carte
    const centerX = rect.width / 2;
    const centerY = rect.height / 2;
    const rotateX = -(y - centerY) / 5;
    const rotateY = (x - centerX) / 5;
    card.style.transform = `rotateX(${rotateX}deg) rotateY(${rotateY}deg)`;
  });

  card.addEventListener('mouseleave', () => {
    card.style.transform = 'rotateX(0deg) rotateY(0deg)';
  });
});


document.querySelectorAll('.card').forEach(card => {
  card.addEventListener('click', e => {
    e.preventDefault(); // empêche la redirection immédiate

    const url = card.getAttribute('href'); // récupère le lien de la card

    card.classList.add('launch'); // ajoute l’effet

    setTimeout(() => {
      window.location.href = url; // redirection après l’animation
    }, 300); // même durée que la transition CSS
  });
});

window.addEventListener('pageshow', () => { // Réinitialise l’effet au retour sur la page 
  document.querySelectorAll('.card').forEach(card => {
    card.classList.remove('launch');
  });
});
