// Système de suggestions pour la recherche
(function() {
    'use strict';

    // Configuration
    const MIN_CHARS = 2;        // Min 2 car pour chercher
    const MAX_RESULTS = 8;      // Max 8 suggestion
    const DEBOUNCE_MS = 200;    // Attendre 200ms après la derniere frappe
    const API_URL = 'https://groupietrackers.herokuapp.com/api/artists';

    // Cache des donnee
    let allArtists = [];
    let suggestionsCache = [];

    // Fonction debounce pour optimiser les performances
    function debounce(fn, delay) {
        let timeout;
        return function(...args) {
            clearTimeout(timeout);
            timeout = setTimeout(() => fn.apply(this, args), delay);
        };
    }

    // Normaliser le texte (minuscules, sans accents)
    function normalize(text) {
        return String(text || '')
            .toLowerCase()
            .normalize('NFD')
            .replace(/[\u0300-\u036f]/g, '')
            .trim();
    }

    // Charger les artistes depuis l'API au démarrage
    async function loadArtists() {
        try {
            console.log('🎸 Chargement des artistes...');
            const response = await fetch(API_URL);

            if (!response.ok) {
                console.error('Erreur API:', response.status);
                return;
            }

            allArtists = await response.json();
            buildSuggestionsCache();
            console.log(' Artistes chargés:', allArtists.length);
        } catch (error) {
            console.error('Erreur chargement artistes:', error);
        }
    }

    // Construire le cache de suggestions
    function buildSuggestionsCache() {
        suggestionsCache = [];

        allArtists.forEach(artist => {
            // Ajouter le nom de l'artiste
            suggestionsCache.push({
                type: 'artist',
                value: artist.name,
                id: artist.id,
                image: artist.image,
                year: artist.creationDate
            });

            // Ajouter les membres
            if (artist.members && Array.isArray(artist.members)) {
                artist.members.forEach(member => {
                    suggestionsCache.push({
                        type: 'member',
                        value: member,
                        id: artist.id,
                        artistName: artist.name,
                        image: artist.image
                    });
                });
            }

            // Ajouter l'année
            suggestionsCache.push({
                type: 'year',
                value: String(artist.creationDate),
                id: artist.id,
                artistName: artist.name,
                image: artist.image
            });
        });

        console.log('✅ Cache construit:', suggestionsCache.length, 'entrées');
    }

    // Filtrer les suggestions selon la requête
    function filterSuggestions(query) {
        if (!query || query.length < MIN_CHARS) {
            return [];
        }

        const normalizedQuery = normalize(query);
        const results = [];
        const seen = new Set();

        // Trier par priorité commence par > contient
        const starts = [];
        const contains = [];

        suggestionsCache.forEach(item => {
            const normalizedValue = normalize(item.value);

            if (normalizedValue.startsWith(normalizedQuery)) {
                starts.push(item);
            } else if (normalizedValue.includes(normalizedQuery)) {
                contains.push(item);
            }
        });

        // Combiner et éviter les doublons
        [...starts, ...contains].forEach(item => {
            const key = `${item.type}-${item.id}-${item.value}`;
            if (!seen.has(key) && results.length < MAX_RESULTS) {
                seen.add(key);
                results.push(item);
            }
        });

        return results;
    }

    // Créer le conteneur de suggestions
    function createSuggestionsBox(input) {
        // Wrapper pour position relative
        const wrapper = document.createElement('div');
        wrapper.className = 'suggestions-wrapper';
        wrapper.style.position = 'relative';
        wrapper.style.width = '100%';

        const parent = input.parentElement;
        parent.insertBefore(wrapper, input);
        wrapper.appendChild(input);

        // Box de suggestions
        const box = document.createElement('div');
        box.className = 'suggestions-box';
        box.style.cssText = `
            position: absolute;
            left: 0;
            right: 0;
            top: calc(100% + 8px);
            z-index: 9999;
            background: var(--bg-secondary, #fff);
            border: 1px solid var(--border-color, rgba(0,0,0,0.1));
            border-radius: 12px;
            box-shadow: 0 8px 24px rgba(0,0,0,0.15);
            max-height: 400px;
            overflow-y: auto;
            display: none;
        `;

        wrapper.appendChild(box);
        return box;
    }

    // Afficher les suggestions
    function showSuggestions(box, input, suggestions) {
        if (!suggestions || suggestions.length === 0) {
            box.style.display = 'none';
            return;
        }

        box.innerHTML = '';

        suggestions.forEach((item, index) => {
            const button = document.createElement('button');
            button.type = 'button';
            button.className = 'suggestion-item';
            button.style.cssText = `
                width: 100%;
                padding: 12px 16px;
                border: none;
                background: transparent;
                text-align: left;
                cursor: pointer;
                display: flex;
                align-items: center;
                gap: 12px;
                color: var(--text-primary, #111);
                transition: background 0.2s;
            `;

            // Image (si disponible)
            if (item.image) {
                const img = document.createElement('img');
                img.src = item.image;
                img.style.cssText = `
                    width: 40px;
                    height: 40px;
                    border-radius: 8px;
                    object-fit: cover;
                `;
                button.appendChild(img);
            }

            // Texte
            const textDiv = document.createElement('div');
            textDiv.style.flex = '1';

            const mainText = document.createElement('div');
            mainText.textContent = item.value;
            mainText.style.fontWeight = '500';
            textDiv.appendChild(mainText);

            // Sous-texte selon le type
            if (item.type === 'member' && item.artistName) {
                const subText = document.createElement('div');
                subText.textContent = `Membre de ${item.artistName}`;
                subText.style.cssText = `
                    font-size: 0.85em;
                    color: var(--text-secondary, #666);
                    margin-top: 2px;
                `;
                textDiv.appendChild(subText);
            } else if (item.type === 'year' && item.artistName) {
                const subText = document.createElement('div');
                subText.textContent = `${item.artistName} - Créé en ${item.value}`;
                subText.style.cssText = `
                    font-size: 0.85em;
                    color: var(--text-secondary, #666);
                    margin-top: 2px;
                `;
                textDiv.appendChild(subText);
            } else if (item.type === 'artist' && item.year) {
                const subText = document.createElement('div');
                subText.textContent = `Groupe créé en ${item.year}`;
                subText.style.cssText = `
                    font-size: 0.85em;
                    color: var(--text-secondary, #666);
                    margin-top: 2px;
                `;
                textDiv.appendChild(subText);
            }

            button.appendChild(textDiv);

            // Hover effect
            button.addEventListener('mouseenter', () => {
                button.style.background = 'var(--bg-primary, rgba(0,0,0,0.05))';
            });
            button.addEventListener('mouseleave', () => {
                button.style.background = 'transparent';
            });

            // Click sur la suggestion
            button.addEventListener('click', () => {
                input.value = item.value;
                box.style.display = 'none';

                // Soumettre le formulaire
                const form = input.closest('form');
                if (form) {
                    form.submit();
                }
            });

            box.appendChild(button);

            // Séparateur (sauf pour le dernier)
            if (index < suggestions.length - 1) {
                const separator = document.createElement('div');
                separator.style.cssText = `
                    height: 1px;
                    background: var(--border-color, rgba(0,0,0,0.08));
                    margin: 0;
                `;
                box.appendChild(separator);
            }
        });

        box.style.display = 'block';
    }

    // Cacher les suggestions
    function hideSuggestions(box) {
        box.style.display = 'none';
    }

    // Initialiser les suggestions sur un input
    function initSuggestionsForInput(input) {
        const box = createSuggestionsBox(input);

        // Fonction de mise à jour avec debounce
        const updateSuggestions = debounce(() => {
            const query = input.value.trim();

            if (query.length < MIN_CHARS) {
                hideSuggestions(box);
                return;
            }

            const suggestions = filterSuggestions(query);
            showSuggestions(box, input, suggestions);
        }, DEBOUNCE_MS);

        // Event listeners
        input.addEventListener('input', updateSuggestions);

        input.addEventListener('focus', () => {
            if (input.value.trim().length >= MIN_CHARS) {
                updateSuggestions();
            }
        });

        // Cacher au clic extérieur
        document.addEventListener('click', (e) => {
            if (!input.contains(e.target) && !box.contains(e.target)) {
                hideSuggestions(box);
            }
        });

        // Cacher au blur (avec délai pour permettre le clic)
        input.addEventListener('blur', () => {
            setTimeout(() => hideSuggestions(box), 200);
        });
    }

    // Initialisation au chargement de la page
    document.addEventListener('DOMContentLoaded', async () => {
        console.log('🎸 Initialisation des suggestions...');

        // Charger les artistes d'abord
        await loadArtists();

        // Trouver tous les inputs de recherche
        const searchInputs = document.querySelectorAll('input.search-input');

        if (searchInputs.length === 0) {
            console.log('Aucun input .search-input trouvé');
            return;
        }

        console.log(`✅ ${searchInputs.length} input(s) trouvé(s)`);

        // Initialiser les suggestions pour chaque input
        searchInputs.forEach(input => {
            initSuggestionsForInput(input);
        });
    });
})();