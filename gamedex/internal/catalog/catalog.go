// Package catalog contient les jeux et les tournois en mémoire.
// Il ne connaît ni les templates, ni HTTP : il reçoit ses données d'une Source.
package catalog

import (
	"fmt"
	"sort"
	"strings"
)

// Game correspond au JSON de la mini API.
type Game struct {
	ID     int      `json:"id"`
	Name   string   `json:"name"`
	Year   int      `json:"year"`
	Studio string   `json:"studio"`
	Genres []string `json:"genres"`
}

// Tournament correspond au JSON de la mini API.
type Tournament struct {
	ID     int     `json:"id"`
	GameID int     `json:"gameId"`
	Name   string  `json:"name"`
	City   string  `json:"city"`
	Date   string  `json:"date"`
	Lat    float64 `json:"lat"`
	Lng    float64 `json:"lng"`
}

// Source est tout ce qui sait fournir des jeux et des tournois (ici : le client de la mini API).
type Source interface {
	FetchGames() ([]Game, error)
	FetchTournaments(gameID int) ([]Tournament, error)
}

// Catalog est chargé une seule fois au démarrage, puis uniquement lu.
type Catalog struct {
	games       []Game
	tournaments map[int][]Tournament
}

// Load interroge la source une seule fois : tous les jeux, puis les tournois de chacun.
func Load(src Source) (*Catalog, error) {
	games, err := src.FetchGames()
	if err != nil {
		return nil, fmt.Errorf("chargement des jeux : %w", err)
	}
	c := &Catalog{games: games, tournaments: make(map[int][]Tournament, len(games))}
	for _, g := range games {
		ts, err := src.FetchTournaments(g.ID)
		if err != nil {
			return nil, fmt.Errorf("chargement des tournois du jeu %d : %w", g.ID, err)
		}
		c.tournaments[g.ID] = ts
	}
	return c, nil
}

func (c *Catalog) Games() []Game { return c.games }

func (c *Catalog) Game(id int) (Game, bool) {
	for _, g := range c.games {
		if g.ID == id {
			return g, true
		}
	}
	return Game{}, false
}

func (c *Catalog) Tournaments(gameID int) []Tournament { return c.tournaments[gameID] }

// Genres renvoie la liste triée, sans doublon, des genres de tous les jeux.
func (c *Catalog) Genres() []string {
	seen := map[string]bool{}
	var out []string
	for _, g := range c.games {
		for _, genre := range g.Genres {
			if !seen[genre] {
				seen[genre] = true
				out = append(out, genre)
			}
		}
	}
	sort.Strings(out)
	return out
}

// Filter garde les jeux dont le nom contient q (sans tenir compte de la casse)
// et qui ont le genre demandé. Un critère vide ne filtre rien.
func (c *Catalog) Filter(q, genre string) []Game {
	q = strings.ToLower(strings.TrimSpace(q))
	genre = strings.ToLower(strings.TrimSpace(genre))

	out := []Game{}
	for _, g := range c.games {
		if q != "" && !strings.Contains(strings.ToLower(g.Name), q) {
			continue
		}
		if genre != "" && !hasGenre(g, genre) {
			continue
		}
		out = append(out, g)
	}
	return out
}

func hasGenre(g Game, genre string) bool {
	for _, x := range g.Genres {
		if strings.ToLower(x) == genre {
			return true
		}
	}
	return false
}
