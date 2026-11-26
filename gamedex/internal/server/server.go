// Package server contient les handlers HTTP et les templates.
// Il lit le catalog ; il n'importe pas le package api.
package server

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"gamedex/internal/catalog"
)

// Searcher : ce que le serveur sait demander à l'extérieur pour la recherche POST (bonus).
// C'est main qui branche le vrai client de l'API ; le serveur ne le connaît pas.
type Searcher interface {
	SearchGames(q, genre string) ([]catalog.Game, error)
}

type Server struct {
	cat     *catalog.Catalog
	search  Searcher
	tpl     *template.Template
	webRoot string
}

// PageData : les données passées aux templates.
type PageData struct {
	Title       string
	Query       string
	Genre       string
	Genres      []string
	Games       []catalog.Game
	Game        catalog.Game
	Tournaments []catalog.Tournament
	Status      int
	Message     string
}

func New(cat *catalog.Catalog, search Searcher, webRoot string) (*Server, error) {
	tpl, err := template.ParseGlob(filepath.Join(webRoot, "templates", "*.html"))
	if err != nil {
		return nil, err
	}
	return &Server{cat: cat, search: search, tpl: tpl, webRoot: webRoot}, nil
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.accueil)
	mux.HandleFunc("GET /jeux", s.listeJeux)
	mux.HandleFunc("GET /jeux/{id}", s.detailJeu)
	mux.HandleFunc("POST /recherche", s.recherche)

	static := http.FileServer(http.Dir(filepath.Join(s.webRoot, "static")))
	mux.Handle("GET /static/", http.StripPrefix("/static/", static))

	mux.HandleFunc("/", s.pageIntrouvable) // toute autre URL : page d'erreur 404
	return mux
}

// render exécute le template dans un buffer avant d'écrire : en cas d'erreur
// de template, on peut encore renvoyer un vrai 500 propre.
func (s *Server) render(w http.ResponseWriter, status int, name string, data PageData) {
	var buf bytes.Buffer
	if err := s.tpl.ExecuteTemplate(&buf, name, data); err != nil {
		log.Printf("erreur de template %s : %v", name, err)
		http.Error(w, "erreur interne", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	buf.WriteTo(w)
}

func (s *Server) erreur(w http.ResponseWriter, status int, message string) {
	s.render(w, status, "erreur.html", PageData{Title: "Erreur", Status: status, Message: message})
}
