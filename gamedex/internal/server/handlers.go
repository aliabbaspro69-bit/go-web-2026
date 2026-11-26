package server

import (
	"log"
	"net/http"
	"strconv"
	"strings"
)

func (s *Server) pageIntrouvable(w http.ResponseWriter, r *http.Request) {
	s.erreur(w, http.StatusNotFound, "Cette page n'existe pas.")
}

func (s *Server) accueil(w http.ResponseWriter, r *http.Request) {
	s.render(w, http.StatusOK, "accueil.html", PageData{Title: "Accueil", Genres: s.cat.Genres()})
}

// GET /jeux?q=...&genre=... : filtre côté Go, sans aucun appel à l'API.
func (s *Server) listeJeux(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	genre := strings.TrimSpace(r.URL.Query().Get("genre"))

	s.render(w, http.StatusOK, "jeux.html", PageData{
		Title:  "Les jeux",
		Query:  q, // html/template échappe le contenu à l'affichage
		Genre:  genre,
		Genres: s.cat.Genres(),
		Games:  s.cat.Filter(q, genre),
	})
}

func (s *Server) detailJeu(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		s.erreur(w, http.StatusBadRequest, "L'identifiant du jeu doit être un nombre.")
		return
	}
	game, ok := s.cat.Game(id)
	if !ok {
		s.erreur(w, http.StatusNotFound, "jeu introuvable")
		return
	}
	s.render(w, http.StatusOK, "jeu.html", PageData{
		Title:       game.Name,
		Game:        game,
		Tournaments: s.cat.Tournaments(id),
	})
}

// POST /recherche (bonus) : le serveur transmet q et genre à la mini API.
func (s *Server) recherche(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<16)
	if err := r.ParseForm(); err != nil {
		s.erreur(w, http.StatusBadRequest, "Formulaire invalide.")
		return
	}
	q := strings.TrimSpace(r.PostFormValue("q"))
	genre := strings.TrimSpace(r.PostFormValue("genre"))

	games, err := s.search.SearchGames(q, genre)
	if err != nil {
		log.Printf("recherche via l'API : %v", err)
		s.erreur(w, http.StatusBadGateway, "La recherche auprès de l'API a échoué : "+err.Error())
		return
	}
	s.render(w, http.StatusOK, "jeux.html", PageData{
		Title:  "Résultats",
		Query:  q,
		Genre:  genre,
		Genres: s.cat.Genres(),
		Games:  games,
	})
}
