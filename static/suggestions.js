(() => {
    "use strict";

    // reglages de base (si ya trop d'resultats c moche)
    const MIN_CHARS = 2;
    const MAX_RESULTS = 8;
    const DEBOUNCE_MS = 150;

    // les liens pr chercher les noms dartistes sur le serv
    const SUGGEST_ENDPOINTS = [
        (q) => `/suggest?q=${encodeURIComponent(q)}`,
        (q) => `/api/suggest?q=${encodeURIComponent(q)}`,
        (q) => `/suggestions?q=${encodeURIComponent(q)}`
    ];

    // Jutilise un debounce : jattends ~150 ms apres la derniere frappe
    // Ca evite de spammer et ca ameliore les perfs du site direct
    const debounce = (fn, ms) => {
        let t;
        return (...args) => {
            clearTimeout(t);
            t = setTimeout(() => fn(...args), ms);
        };
    };

    // pr clean le texte (vire les accents + minuscule)
    const normalize = (s) =>
        String(s || "")
            .toLowerCase()
            .normalize("NFD")
            .replace(/[\u0300-\u036f]/g, "")
            .trim();

    const uniq = (arr) => Array.from(new Set(arr.filter(Boolean)));

    // la je trie ce que le serv me renvoie pr pas avoir d'bugs
    const parseSuggestPayload = (data) => {
        if (!data) return [];
        if (Array.isArray(data)) {
            if (typeof data[0] === "string") return data;
            return data
                .map((x) => x?.name || x?.Name || x?.title || x?.Title || "")
                .filter(Boolean);
        }
        const list =
            data.results || data.Results || data.items || data.Items || data.suggestions || data.Suggestions || [];
        if (Array.isArray(list)) return parseSuggestPayload(list);
        return [];
    };

    // Je recupere les noms dartistes deja presents ds la page (h1, h2, etc)
    const collectLocalArtistNames = () => {
        const names = [];
        document.querySelectorAll(".artist-name").forEach((el) => {
            const t = el.textContent?.trim();
            if (t) names.push(t);
        });
        document.querySelectorAll("select option").forEach((opt) => {
            const txt = opt.textContent?.trim();
            if (txt && !txt.startsWith("--")) {
                const cleaned = txt.replace(/\s*\(\d{4}\)\s*$/, "").trim();
                if (cleaned) names.push(cleaned);
            }
        });
        document.querySelectorAll("h1, h2").forEach((h) => {
            const t = h.textContent?.trim();
            if (t && t.length <= 40 && !/nos artistes|résultats|compar/i.test(normalize(t))) {
                names.push(t);
            }
        });
        return uniq(names);
    };

    // fct pr aller chercher les suggestions sur le web (fetch)
    const tryFetchSuggestions = async (query) => {
        for (const makeUrl of SUGGEST_ENDPOINTS) {
            const url = makeUrl(query);
            try {
                const res = await fetch(url, { headers: { "Accept": "application/json" } });
                if (!res.ok) continue;
                const ct = res.headers.get("content-type") || "";
                if (!ct.includes("application/json")) continue;
                const json = await res.json();
                const parsed = parseSuggestPayload(json);
                if (parsed.length) return parsed;
            } catch { }
        }
        return null;
    };

    // Je filtre ceux qui commencent par ou contiennent le texte (local)
    const localSuggest = (query, localList) => {
        const qn = normalize(query);
        if (!qn) return [];
        const starts = [];
        const includes = [];
        for (const name of localList) {
            const nn = normalize(name);
            if (!nn) continue;
            if (nn.startsWith(qn)) starts.push(name);
            else if (nn.includes(qn)) includes.push(name);
        }
        return [...starts, ...includes].slice(0, MAX_RESULTS);
    };

    // Je cree une ptite box en HTML juste en dessous de la barre
    const ensureDropdown = (input) => {
        const wrap = document.createElement("div");
        wrap.className = "suggestions-wrap";
        wrap.style.position = "relative";
        wrap.style.width = "100%";

        const parent = input.parentElement;
        parent.insertBefore(wrap, input);
        wrap.appendChild(input);

        // J'affiche un dropdown sous l'input direct
        const box = document.createElement("div");
        box.className = "suggestions-box";
        box.style.position = "absolute";
        box.style.left = "0";
        box.style.right = "0";
        box.style.top = "calc(100% + 8px)";
        box.style.zIndex = "9999";
        box.style.borderRadius = "14px";
        box.style.overflow = "hidden";
        box.style.display = "none";
        box.style.background = "var(--bg-secondary, #fff)";
        box.style.border = "1px solid var(--border-color, rgba(0,0,0,0.12))";
        box.style.boxShadow = "0 10px 30px rgba(0,0,0,0.15)";

        wrap.appendChild(box);
        return box;
    };

    const hideBox = (box) => {
        box.style.display = "none";
        box.innerHTML = "";
    };

    // Je genere une liste de suggestions ds le HTML (boucle)
    const renderBox = (box, input, items) => {
        if (!items || !items.length) {
            hideBox(box);
            return;
        }

        box.innerHTML = "";
        items.forEach((txt, idx) => {
            const row = document.createElement("button");
            row.type = "button";
            row.className = "suggestion-item";
            row.textContent = txt;
            row.style.width = "100%";
            row.style.textAlign = "left";
            row.style.padding = "12px 14px";
            row.style.border = "0";
            row.style.background = "transparent";
            row.style.cursor = "pointer";
            row.style.color = "var(--text-primary, #111)";

            row.addEventListener("mouseenter", () => {
                row.style.background = "var(--bg-primary, rgba(0,0,0,0.05))";
            });
            row.addEventListener("mouseleave", () => {
                row.style.background = "transparent";
            });

            // Qd on clique une suggestion, je remplis linput et je lance la rech
            row.addEventListener("click", () => {
                input.value = txt;
                hideBox(box);
                const form = input.closest("form");
                if (form) form.submit();
            });

            box.appendChild(row);

            if (idx !== items.length - 1) {
                const sep = document.createElement("div");
                sep.style.height = "1px";
                sep.style.background = "var(--border-color, rgba(0,0,0,0.08))";
                box.appendChild(sep);
            }
        });

        box.style.display = "block";
    };

    // --- DEBUT DU SCRIPT ---
    document.addEventListener("DOMContentLoaded", () => {
        // Dans suggestions.js, je met un listener sur linput .search-input
        const inputs = Array.from(document.querySelectorAll("input.search-input"));
        if (!inputs.length) return;

        const localNames = collectLocalArtistNames();

        for (const input of inputs) {
            const box = ensureDropdown(input);

            // A chaque frappe, je recupere la valeur (avec le debounce)
            // Je lance pas une requete a chaque lettre sinon ca lag !
            const update = debounce(async () => {
                const q = input.value.trim();
                if (q.length < MIN_CHARS) return hideBox(box);

                // Je cherche dabord sur le serv, sinon je prend ma liste locale
                const server = await tryFetchSuggestions(q);
                if (server && server.length) {
                    renderBox(box, input, uniq(server).slice(0, MAX_RESULTS));
                    return;
                }

                const local = localSuggest(q, localNames);
                renderBox(box, input, local);
            }, DEBOUNCE_MS);

            // J'ecoute ce que lutilisateur tape
            input.addEventListener("input", update);

            // Si on reclique ds la barre on raffiche les suggestions
            input.addEventListener("focus", () => {
                if (input.value.trim().length >= MIN_CHARS) update();
            });

            // Gestion des touches (haut, bas, enter)
            input.addEventListener("keydown", (e) => {
                if (box.style.display === "none") return;
                const buttons = Array.from(box.querySelectorAll("button.suggestion-item"));
                if (!buttons.length) return;
                const active = box.querySelector("button.suggestion-item[data-active='1']");
                let idx = active ? buttons.indexOf(active) : -1;

                if (e.key === "ArrowDown") {
                    e.preventDefault();
                    idx = Math.min(idx + 1, buttons.length - 1);
                } else if (e.key === "ArrowUp") {
                    e.preventDefault();
                    idx = Math.max(idx - 1, 0);
                } else if (e.key === "Enter") {
                    if (active) {
                        e.preventDefault();
                        active.click();
                    }
                    return;
                } else if (e.key === "Escape") {
                    hideBox(box);
                    return;
                } else { return; }

                buttons.forEach((b) => b.removeAttribute("data-active"));
                const next = buttons[idx];
                next.setAttribute("data-active", "1");
                next.style.background = "var(--bg-primary, rgba(0,0,0,0.05))";
                next.scrollIntoView({ block: "nearest" });
                buttons.forEach((b, i) => { if (i !== idx) b.style.background = "transparent"; });
            });

            // Fermer la box si on clique a coter (important ca)
            document.addEventListener("click", (e) => {
                if (!box.contains(e.target) && e.target !== input) hideBox(box);
            });

            input.addEventListener("blur", () => {
                setTimeout(() => hideBox(box), 150);
            });
        }
    });
})();