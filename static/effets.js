
window.SearchBar = window.SearchBar || {};

console.log(" SearchBar script loading ...");

// Configuration adaptative
window.SearchBar.config = {
  searchInputSelector: "#searchBar, #searchInput",
  suggestionsSelector: "#suggestions",
  cardsSelector: ".grid .card, .artistsGrid .card",
  suggestEndpoint: "/suggest",
  searchEndpoint: "/search",
  maxSuggestions: 10,
  debounceDelay: 150
};

// sert a garder l'état de la recherche et des suggestions
window.SearchBar.state = {
  lastQuery: '',
  debounceTimer: null,
  isInitialized: false
};

// animation et effets
window.SearchBar.animations = {
  // Animation de fade in pour les suggestions
  fadeInSuggestions(element) {
    if (!element) return;
    element.classList.remove("hidden");
    element.classList.add("active");
    console.log("✓ Suggestions animées - fadeIn");
  },

  // Animation de fade out pour les suggestions
  fadeOutSuggestions(element) {
    if (!element) return;
    element.classList.remove("active");
    element.classList.add("hidden");
    console.log("✓ Suggestions animées - fadeOut");
  },

  // Animation de card au hover
  addCardHoverAnimation(card) {
    if (!card) return;
    card.style.transition = "all 0.4s cubic-bezier(0.34, 1.56, 0.64, 1)";
  },

  // Animation de pulse
  pulseElement(element) {
    if (!element) return;
    element.style.animation = "pulse 0.6s ease-in-out";
  },

  // Highlight d'une carte
  highlightCard(card) {
    if (!card) return;
    card.style.animation = "pulse 0.8s ease-in-out";
    card.style.boxShadow = "0 0 30px rgba(59, 130, 246, 0.7)";
    setTimeout(() => {
      card.style.boxShadow = "0 6px 18px rgba(0,0,0,0.22)";
      card.style.animation = "none";
    }, 800);
  },

  // Animer les cartes au chargement
  animateCardsOnLoad(cards) {
    if (!cards || cards.length === 0) {
      console.warn("⚠ Aucune carte trouvée pour l'animation");
      return;
    }
  
    cards.forEach((card, index) => {
      
      card.style.animation = 'none';
      
      void card.offsetWidth;
      
      card.style.animation = `cardBounce 0.6s cubic-bezier(0.34, 1.56, 0.64, 1) ${index * 80}ms forwards`;
      
      
      card.addEventListener("mouseenter", () => window.SearchBar.animations.createBorderStripes(card));
      card.addEventListener("mouseleave", () => window.SearchBar.animations.removeBorderStripes(card));
    });
    
    console.log(`✓ ${cards.length} cartes animées au chargement`);
  },

  
  createBorderStripes(element) {
    if (!element) return;
    
    
    const isInput = element.tagName === "INPUT";
    const parentElement = isInput ? element.parentElement : element;
    
    if (parentElement.querySelector(".card-stripes")) return; 
    
    
    const computedStyle = window.getComputedStyle(element);
    if (computedStyle.position !== "relative" && computedStyle.position !== "absolute") {
      element.style.position = "relative";
    }
    
    const stripesContainer = document.createElement("div");
    stripesContainer.className = "card-stripes";
    stripesContainer.style.borderRadius = computedStyle.borderRadius;
    
    
    if (isInput) {
      stripesContainer.style.position = "absolute";
      stripesContainer.style.top = element.offsetTop + "px";
      stripesContainer.style.left = element.offsetLeft + "px";
      stripesContainer.style.width = element.offsetWidth + "px";
      stripesContainer.style.height = element.offsetHeight + "px";
      stripesContainer.style.pointerEvents = "none";
      stripesContainer.style.zIndex = "10";
    }
    
    
    const stripeTypes = ["stripe-top", "stripe-right", "stripe-bottom", "stripe-left"];
    stripeTypes.forEach(type => {
      const stripe = document.createElement("div");
      stripe.className = `stripe ${type}`;
      stripesContainer.appendChild(stripe);
    });
    
    parentElement.appendChild(stripesContainer);
    console.log(" Raies blanches tournantes créées");
  },

  
  removeBorderStripes(element) {
    if (!element) return;
    const isInput = element.tagName === "INPUT";
    const parentElement = isInput ? element.parentElement : element;
    const stripes = parentElement.querySelector(".card-stripes");
    if (stripes) {
      stripes.remove();
    }
    console.log("✓ Raies blanches supprimées");
  }
};


window.SearchBar.utils = {
  
  getSearchInput() {
    const el = document.querySelector(window.SearchBar.config.searchInputSelector);
    if (!el) console.warn(" Input de recherche non trouvé");
    return el;
  },

  
  getSuggestionsBox() {
    const el = document.querySelector(window.SearchBar.config.suggestionsSelector);
    if (!el) console.warn(" Box suggestions non trouvé");
    return el;
  },

  
  getCards() {
    const els = document.querySelectorAll(window.SearchBar.config.cardsSelector);
    if (els.length === 0) console.warn("⚠ Aucune carte trouvée");
    return els;
  },

  
  getArtistName(item) {
    return item.name || item.Name || "";
  },

  
  getArtistId(item) {
    return item.id || item.ID || "";
  },

  
  applyAnimationToCards() {
    const cards = window.SearchBar.utils.getCards();
    cards.forEach(card => {
      window.SearchBar.animations.addCardHoverAnimation(card);
    });
  }
};


window.SearchBar.updateSuggestions = async function(query) {
  const input = window.SearchBar.utils.getSearchInput();
  if (!input) return;

  const box = window.SearchBar.utils.getSuggestionsBox();
  if (!box) return;

  if (query.length < 1) {
    window.SearchBar.animations.fadeOutSuggestions(box);
    box.innerHTML = "";
    return;
  }

  try {
    const res = await fetch(`${window.SearchBar.config.suggestEndpoint}?q=${encodeURIComponent(query)}`);
    
    if (!res.ok) throw new Error("API error");
    
    let items = await res.json();
    if (!Array.isArray(items)) items = [];

    
    const filtered = items.filter(item =>
      window.SearchBar.utils.getArtistName(item).toLowerCase().startsWith(query.toLowerCase())
    );

    if (filtered.length > 0) {
      const html = filtered
        .slice(0, window.SearchBar.config.maxSuggestions)
        .map(item => {
          const name = window.SearchBar.utils.getArtistName(item);
          const id = window.SearchBar.utils.getArtistId(item);
          return `<a class="sugg-item" data-id="${id}" href="/artist/${id}">${name}</a>`;
        })
        .join("");
      
      box.innerHTML = html;
      window.SearchBar.animations.fadeInSuggestions(box);
    } else {
      window.SearchBar.animations.fadeOutSuggestions(box);
      box.innerHTML = "";
    }
  } catch (err) {
    console.error(" Erreur recherche:", err);
    window.SearchBar.animations.fadeOutSuggestions(box);
    box.innerHTML = "";
  }
};


window.SearchBar.filterArtists = function(query = "") {
  const input = window.SearchBar.utils.getSearchInput();
  const q = query || (input ? input.value.trim().toLowerCase() : "");
  
  const cards = window.SearchBar.utils.getCards();
  let visibleCount = 0;
  
  cards.forEach((card) => {
    const name = (card.querySelector("h2")?.textContent || "").trim().toLowerCase();
    const shouldShow = !q || name.startsWith(q);
    
    if (shouldShow) {
      card.classList.remove("hidden-search");
      card.style.display = "";
      card.style.animation = `cardBounce 0.6s cubic-bezier(0.34, 1.56, 0.64, 1) ${visibleCount * 80}ms forwards`;
      
      void card.offsetWidth;
      visibleCount++;
    } else {
      card.classList.add("hidden-search");
      card.style.display = "none";
      card.style.animation = "none";
    }
  });
  
  console.log(`✓ Filtrage: ${visibleCount}/${cards.length} cartes affichées`);
};


window.SearchBar.closeSuggestions = function() {
  const box = window.SearchBar.utils.getSuggestionsBox();
  if (box) {
    window.SearchBar.animations.fadeOutSuggestions(box);
    box.innerHTML = "";
  }
};


window.SearchBar.actions = {
  
  searchArtist(artistName) {
    const input = window.SearchBar.utils.getSearchInput();
    if (input) {
      input.value = artistName;
      window.SearchBar.updateSuggestions(artistName);
      window.SearchBar.filterArtists(artistName);
    }
  },

  
  reset() {
    const input = window.SearchBar.utils.getSearchInput();
    if (input) {
      input.value = "";
      window.SearchBar.closeSuggestions();
      window.SearchBar.filterArtists("");
    }
  },

  
  showAll() {
    window.SearchBar.utils.getCards().forEach(card => {
      card.style.display = "";
      card.classList.remove("hidden-search");
    });
  },

  
  highlightArtist(artistName) {
    window.SearchBar.utils.getCards().forEach(card => {
      const name = card.querySelector("h2")?.textContent || "";
      if (name.toLowerCase().includes(artistName.toLowerCase())) {
        window.SearchBar.animations.highlightCard(card);
      }
    });
  }
};


window.SearchBar.init = function() {
  if (window.SearchBar.state.isInitialized) {
    console.log(" SearchBar déjà initialisé");
    return;
  }
  
  console.log("Initialisation SearchBar...");
  
  const input = window.SearchBar.utils.getSearchInput();
  const box = window.SearchBar.utils.getSuggestionsBox();
  const cards = window.SearchBar.utils.getCards();

  if (!input) {
    console.error(" Input de recherche non trouvé");
    return;
  }

  if (cards.length === 0) {
    console.warn(" Aucune carte trouvée");
  } else {
    console.log(` ${cards.length} cartes trouvées`);
  }

  // Appliquer les animations aux cartes
  window.SearchBar.utils.applyAnimationToCards();
  
  // Animer les cartes au chargement
  window.SearchBar.animations.animateCardsOnLoad(cards);
  
  // Appliquer les raies aux boutons de la searchbar
  const searchButtons = document.querySelectorAll(".searchbar button, .searchbar .reset");
  searchButtons.forEach(btn => {
    btn.addEventListener("mouseenter", () => window.SearchBar.animations.createBorderStripes(btn));
    btn.addEventListener("mouseleave", () => window.SearchBar.animations.removeBorderStripes(btn));
  });
  
  // Appliquer les raies aux inputs avec mouseover (pas focus)
  const allInputs = document.querySelectorAll("input[type='search'], input[type='number'], input[type='date']");
  allInputs.forEach(inp => {
    inp.addEventListener("mouseenter", () => window.SearchBar.animations.createBorderStripes(inp));
    inp.addEventListener("mouseleave", () => window.SearchBar.animations.removeBorderStripes(inp));
  });
  
  // Appliquer raies aux liens de détail des cartes
  const cardLinks = document.querySelectorAll(".card a");
  cardLinks.forEach(link => {
    link.addEventListener("mouseenter", () => window.SearchBar.animations.createBorderStripes(link));
    link.addEventListener("mouseleave", () => window.SearchBar.animations.removeBorderStripes(link));
  });
  
  // Appliquer raies aux boutons de filtres
  const filterButtons = document.querySelectorAll(".filters__actions button, .filters__actions .reset");
  filterButtons.forEach(btn => {
    btn.addEventListener("mouseenter", () => window.SearchBar.animations.createBorderStripes(btn));
    btn.addEventListener("mouseleave", () => window.SearchBar.animations.removeBorderStripes(btn));
  });

  // Événement input avec debounce
  input.addEventListener("input", () => {
    const q = input.value.trim();
    if (q === window.SearchBar.state.lastQuery) return;

    window.SearchBar.state.lastQuery = q;
    clearTimeout(window.SearchBar.state.debounceTimer);

    if (!q) {
      window.SearchBar.closeSuggestions();
      window.SearchBar.filterArtists("");
      return;
    }

    window.SearchBar.state.debounceTimer = setTimeout(() => {
      window.SearchBar.updateSuggestions(q);
      window.SearchBar.filterArtists(q);
    }, window.SearchBar.config.debounceDelay);
  });

  // Événement Enter
  input.addEventListener("keydown", (e) => {
    if (e.key === "Enter") {
      e.preventDefault();
      window.SearchBar.closeSuggestions();
      window.SearchBar.filterArtists(input.value.trim());
    }
  });

  // Clic sur les suggestions
  document.addEventListener("click", (e) => {
    if (e.target.classList.contains("sugg-item")) {
      input.value = e.target.textContent;
      window.SearchBar.closeSuggestions();
      window.SearchBar.filterArtists(e.target.textContent);
    } else if (box && !box.contains(e.target) && e.target !== input) {
      window.SearchBar.closeSuggestions();
    }
  });

  window.SearchBar.state.isInitialized = true;
  console.log(" SearchBar initialized complètement");
  console.log(" API disponible:", Object.keys(window.SearchBar));
};


function initSearchBar() {
  if (document.readyState === "loading") {
    console.log(" DOM en chargement, attendu...");
    document.addEventListener("DOMContentLoaded", () => {
      console.log(" DOM chargé!");
      window.SearchBar.init();
    });
  } else {
    console.log(" DOM déjà chargé!");
    window.SearchBar.init();
  }
}

// Initialisation immédiate
initSearchBar();

