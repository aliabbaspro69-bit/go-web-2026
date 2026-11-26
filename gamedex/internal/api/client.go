// Package api est le client HTTP de la mini API. C'est le seul package qui "parle" à l'API.
package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"gamedex/internal/catalog"
)

const BaseURL = "http://localhost:8081"

const timeout = 3 * time.Second

// Client appelle la mini API avec un timeout.
type Client struct {
	// GamesPath : chemin de la liste des jeux (changez-le pour tester /api/broken ou /api/slow).
	GamesPath string
	http      *http.Client
}

func New() *Client {
	return &Client{GamesPath: "/api/games", http: &http.Client{Timeout: timeout}}
}

// Error est une erreur claire, avec le code HTTP correspondant.
type Error struct {
	Status  int
	Message string
}

func (e *Error) Error() string { return e.Message }

func (c *Client) FetchGames() ([]catalog.Game, error) {
	var games []catalog.Game
	if err := c.get(c.GamesPath, &games); err != nil {
		return nil, err
	}
	return games, nil
}

func (c *Client) FetchTournaments(gameID int) ([]catalog.Tournament, error) {
	var ts []catalog.Tournament
	if err := c.get(fmt.Sprintf("/api/games/%d/tournaments", gameID), &ts); err != nil {
		return nil, err
	}
	return ts, nil
}

// SearchGames appelle /api/games?q=...&genre=... (bonus : recherche transmise à la mini API).
func (c *Client) SearchGames(q, genre string) ([]catalog.Game, error) {
	v := url.Values{}
	if q != "" {
		v.Set("q", q)
	}
	if genre != "" {
		v.Set("genre", genre)
	}
	path := c.GamesPath
	if len(v) > 0 {
		path += "?" + v.Encode()
	}
	var games []catalog.Game
	if err := c.get(path, &games); err != nil {
		return nil, err
	}
	return games, nil
}

func (c *Client) get(path string, out any) error {
	resp, err := c.http.Get(BaseURL + path)
	if err != nil {
		return networkError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return statusError(resp)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		var ne net.Error
		if errors.As(err, &ne) && ne.Timeout() {
			return timeoutError()
		}
		return &Error{http.StatusBadGateway, "la réponse de l'API est illisible"}
	}
	return nil
}

func timeoutError() *Error {
	return &Error{http.StatusGatewayTimeout,
		fmt.Sprintf("l'API met trop de temps à répondre (délai de %s dépassé)", timeout)}
}

func networkError(err error) *Error {
	var ne net.Error
	if errors.As(err, &ne) && ne.Timeout() {
		return timeoutError()
	}
	return &Error{http.StatusBadGateway,
		fmt.Sprintf("API indisponible : impossible de joindre %s (est-elle lancée avec « go run ./api » ?)", BaseURL)}
}

func statusError(resp *http.Response) *Error {
	var body struct {
		Error string `json:"error"`
	}
	_ = json.NewDecoder(io.LimitReader(resp.Body, 1<<16)).Decode(&body)
	msg := body.Error
	if msg == "" {
		msg = http.StatusText(resp.StatusCode)
	}

	switch resp.StatusCode {
	case http.StatusNotFound, http.StatusBadRequest:
		return &Error{resp.StatusCode, msg}
	default:
		return &Error{http.StatusBadGateway,
			fmt.Sprintf("l'API a répondu avec une erreur %d : %s", resp.StatusCode, msg)}
	}
}
