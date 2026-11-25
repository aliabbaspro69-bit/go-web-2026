// Package api contient le client HTTP de la mini API (utilisé par le TP3 et le TP4).
package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

const BaseURL = "http://localhost:8081"

// GamesPath est le chemin de la liste des jeux.
// TP3 question 5 : changez-le en "/api/broken" puis "/api/slow" (ou utilisez le flag -endpoint).
var GamesPath = "/api/games"

const timeout = 3 * time.Second

// Un client avec timeout : sans lui, un serveur lent bloquerait le programme indéfiniment.
var client = &http.Client{Timeout: timeout}

// Game correspond au JSON renvoyé par /api/games et /api/games/{id}.
type Game struct {
	ID     int      `json:"id"`
	Name   string   `json:"name"`
	Year   int      `json:"year"`
	Studio string   `json:"studio"`
	Genres []string `json:"genres"`
}

// Tournament correspond au JSON renvoyé par /api/games/{id}/tournaments.
type Tournament struct {
	ID     int     `json:"id"`
	GameID int     `json:"gameId"`
	Name   string  `json:"name"`
	City   string  `json:"city"`
	Date   string  `json:"date"`
	Lat    float64 `json:"lat"`
	Lng    float64 `json:"lng"`
}

// Error est une erreur claire, avec le code HTTP que GameDex doit renvoyer à son visiteur.
type Error struct {
	Status  int
	Message string
}

func (e *Error) Error() string { return e.Message }

func FetchGames() ([]Game, error) {
	var games []Game
	if err := get(GamesPath, &games); err != nil {
		return nil, err
	}
	return games, nil
}

func FetchGame(id int) (*Game, error) {
	var g Game
	if err := get(fmt.Sprintf("/api/games/%d", id), &g); err != nil {
		return nil, err
	}
	return &g, nil
}

func FetchTournaments(gameID int) ([]Tournament, error) {
	var ts []Tournament
	if err := get(fmt.Sprintf("/api/games/%d/tournaments", gameID), &ts); err != nil {
		return nil, err
	}
	return ts, nil
}

func get(path string, out any) error {
	resp, err := client.Get(BaseURL + path)
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
