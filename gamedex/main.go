package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// Petite "base de données" en mémoire : l'id d'un jeu = sa position + 1.
var jeux = []string{
	"The Legend of Zelda: Breath of the Wild",
	"Celeste",
	"Hollow Knight",
	"Mario Kart 8 Deluxe",
	"Rocket League",
}

func main() {
	port := flag.Int("port", 8080, "port d'écoute du serveur")
	flag.Parse()

	mux := http.NewServeMux()

	// "{$}" = uniquement "/" (sinon "/" attrape toutes les URL inconnues)
	mux.HandleFunc("GET /{$}", accueil)
	mux.HandleFunc("GET /jeux", listeJeux)
	mux.HandleFunc("GET /jeux/{id}", detailJeu)

	// Fichiers statiques : /static/css/style.css -> ./static/css/style.css
	fichiers := http.FileServer(http.Dir("static"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fichiers))

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("GameDex démarre sur http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

// GET / : page HTML qui utilise le CSS (bonus)
func accueil(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `<!DOCTYPE html>
<html lang="fr">
<head>
  <meta charset="utf-8">
  <title>GameDex</title>
  <link rel="stylesheet" href="/static/css/style.css">
</head>
<body>
  <h1>Bienvenue sur GameDex</h1>
  <p>Votre encyclopédie de jeux vidéo.</p>
  <p><a href="/jeux">Voir la liste des jeux</a></p>
</body>
</html>`)
}

// GET /jeux : liste en texte brut
func listeJeux(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	for i, nom := range jeux {
		fmt.Fprintf(w, "%d. %s\n", i+1, nom)
	}
}

// GET /jeux/{id} : 400 si l'id n'est pas un nombre, 404 si le jeu n'existe pas
func detailJeu(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "id invalide : un nombre est attendu", http.StatusBadRequest)
		return
	}
	if id < 1 || id > len(jeux) {
		http.Error(w, "jeu introuvable", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "Jeu n°%d : %s\n", id, jeux[id-1])
}
