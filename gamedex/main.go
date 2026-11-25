package main

import (
	"bytes"
	"errors"
	"flag"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"gamedex/internal/api"
)

var tpl *template.Template

// PageData : les données passées aux templates.
type PageData struct {
	Title       string
	Query       string
	Games       []api.Game
	Game        *api.Game
	Tournaments []api.Tournament
	Status      int
	Message     string
}

func main() {
	port := flag.Int("port", 8080, "port d'écoute du serveur")
	flag.Parse()

	var err error
	tpl, err = template.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("templates introuvables (lancez le programme depuis le dossier gamedex) : %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", accueil)
	mux.HandleFunc("GET /jeux", listeJeux)
	mux.HandleFunc("GET /jeux/{id}", detailJeu)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/", pageIntrouvable) // toute autre URL : page d'erreur 404

	addr := ":" + strconv.Itoa(*port)
	log.Printf("GameDex démarre sur http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

// render exécute le template dans un buffer avant d'écrire : en cas d'erreur
// de template, on peut encore renvoyer un vrai 500 propre.
func render(w http.ResponseWriter, status int, name string, data PageData) {
	var buf bytes.Buffer
	if err := tpl.ExecuteTemplate(&buf, name, data); err != nil {
		log.Printf("erreur de template %s : %v", name, err)
		http.Error(w, "erreur interne", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	buf.WriteTo(w)
}

func erreur(w http.ResponseWriter, status int, message string) {
	render(w, status, "erreur.html", PageData{Title: "Erreur", Status: status, Message: message})
}

// erreurAPI transforme une erreur du client API en page d'erreur avec le bon code.
func erreurAPI(w http.ResponseWriter, err error) {
	var apiErr *api.Error
	if errors.As(err, &apiErr) {
		log.Printf("erreur API : %v", err)
		erreur(w, apiErr.Status, apiErr.Message)
		return
	}
	log.Printf("erreur inattendue : %v", err)
	erreur(w, http.StatusInternalServerError, "erreur interne")
}

func pageIntrouvable(w http.ResponseWriter, r *http.Request) {
	erreur(w, http.StatusNotFound, "Cette page n'existe pas.")
}

func accueil(w http.ResponseWriter, r *http.Request) {
	render(w, http.StatusOK, "accueil.html", PageData{Title: "Accueil"})
}

func listeJeux(w http.ResponseWriter, r *http.Request) {
	games, err := api.FetchGames()
	if err != nil {
		erreurAPI(w, err)
		return
	}
	render(w, http.StatusOK, "jeux.html", PageData{
		Title: "Les jeux",
		Query: r.URL.Query().Get("q"), // affiché tel quel : html/template échappe le contenu
		Games: games,
	})
}

func detailJeu(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		erreur(w, http.StatusBadRequest, "L'identifiant du jeu doit être un nombre.")
		return
	}
	game, err := api.FetchGame(id)
	if err != nil {
		erreurAPI(w, err)
		return
	}
	tournaments, err := api.FetchTournaments(id)
	if err != nil {
		erreurAPI(w, err)
		return
	}
	render(w, http.StatusOK, "jeu.html", PageData{
		Title:       game.Name,
		Game:        game,
		Tournaments: tournaments,
	})
}
