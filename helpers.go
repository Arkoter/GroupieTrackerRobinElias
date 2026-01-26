package main

import (
	"html/template"
	"net/http"
	"sync"
	"time"
)

// structure du cache avec le systeme de rafraichissement
type Cache struct {
	artists         []Artist
	relations       map[int]map[string][]string
	lastRefresh     time.Time
	mutex           sync.RWMutex // pr eviter que le cache plante si 2 pbs y touchent
	refreshInterval time.Duration
}

// jinit le cache a 5 min par defaut
var apiCache = &Cache{
	relations:       make(map[int]map[string][]string),
	refreshInterval: 5 * time.Minute,
}

// fct pr tout clean qd on veut forcer le refresh
func (c *Cache) RefreshCache() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.artists = nil
	c.relations = make(map[int]map[string][]string)
	c.lastRefresh = time.Now()
	return nil
}

// jcheck si ca fait + de 5 min quon a rien recup
func (c *Cache) ShouldRefresh() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return time.Since(c.lastRefresh) > c.refreshInterval
}

// jrecup les artistes : soit ds le cache, soit jappelle lapi
func (c *Cache) GetArtists() ([]Artist, error) {
	c.mutex.RLock()
	// si cest deja la et pas trop vieux jle renvoie direct
	if c.artists != nil && !c.ShouldRefresh() {
		artists := c.artists
		c.mutex.RUnlock()
		return artists, nil
	}
	c.mutex.RUnlock()

	c.mutex.Lock() // je lock pr etre sur que jecris tt seul
	defer c.mutex.Unlock()

	if c.artists != nil && !c.ShouldRefresh() {
		return c.artists, nil
	}

	// la jappelle lapi pcq j'ai rien ds le cache
	artists, err := fetchArtistsFromAPI()
	if err != nil {
		return nil, err
	}

	c.artists = artists
	c.lastRefresh = time.Now()
	return artists, nil
}

// pareil pr les relations mais par ID d'artiste
func (c *Cache) GetRelation(id int) (map[string][]string, error) {
	c.mutex.RLock()
	if relation, ok := c.relations[id]; ok {
		c.mutex.RUnlock()
		return relation, nil
	}
	c.mutex.RUnlock()

	c.mutex.Lock()
	defer c.mutex.Unlock()

	if relation, ok := c.relations[id]; ok {
		return relation, nil
	}

	relation, err := fetchRelationFromAPI(id)
	if err != nil {
		return map[string][]string{}, nil
	}

	c.relations[id] = relation
	return relation, nil
}

// pr charger les fichiers html
func parseTemplate(filepath string) (*template.Template, error) {
	return template.ParseFiles(filepath)
}

// pr l'affichage des stats ds le panel admin
type CacheStatsData struct {
	ArtistsCount   int
	RelationsCount int
	LastRefresh    string
	NextRefresh    string
	CacheAge       string
}

// jcalcule l'age du cache et tt le reste pr l'utilisateur
func GetCacheStats() CacheStatsData {
	apiCache.mutex.RLock()
	defer apiCache.mutex.RUnlock()

	age := time.Since(apiCache.lastRefresh)
	nextRefresh := apiCache.refreshInterval - age
	if nextRefresh < 0 {
		nextRefresh = 0
	}

	return CacheStatsData{
		ArtistsCount:   len(apiCache.artists),
		RelationsCount: len(apiCache.relations),
		LastRefresh:    apiCache.lastRefresh.Format("15:04:05"),
		NextRefresh:    formatDuration(nextRefresh),
		CacheAge:       formatDuration(age),
	}
}

// pr afficher les secondes ou minutes proprement
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return d.Round(time.Second).String()
	}
	return d.Round(time.Minute).String()
}

// le handler pr voir letat du cache sur une page web
func CacheInfoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		renderError(w, r, http.StatusMethodNotAllowed, "Metehode pas autorisee", "Seulement GET")
		return
	}

	stats := GetCacheStats()
	tmpl, err := parseTemplate("templates/cache_info.html")
	if err != nil {
		renderError(w, r, http.StatusInternalServerError, "Err serv", "Impossible dcharger la page.")
		return
	}

	data := PageData{
		Theme: getThemeClass(r),
		Data:  stats,
	}
	_ = tmpl.Execute(w, data)
}

// le handler pr forcer l'update du cache via un bouton
func RefreshCacheHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		renderError(w, r, http.StatusMethodNotAllowed, "Methode pas autorisee", "Faut du POST")
		return
	}

	err := apiCache.RefreshCache()
	if err != nil {
		renderError(w, r, http.StatusInternalServerError, "Err", "Ca veut pas refresh")
		return
	}

	// une fois fini jredirige vers la page d'info
	http.Redirect(w, r, "/cache-info", http.StatusSeeOther)
}
