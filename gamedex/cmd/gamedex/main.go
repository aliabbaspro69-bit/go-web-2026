// GameDex : le site web. Lancement (depuis le dossier gamedex) : go run ./cmd/gamedex
package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"

	"gamedex/internal/api"
	"gamedex/internal/catalog"
	"gamedex/internal/server"
)

func main() {
	port := flag.Int("port", 8080, "port d'écoute du serveur")
	web := flag.String("web", "web", "dossier contenant templates/ et static/")
	flag.Parse()

	client := api.New()

	// Les jeux et les tournois sont chargés UNE seule fois, ici, au démarrage.
	cat, err := catalog.Load(client)
	if err != nil {
		log.Fatalf("démarrage impossible : %v", err)
	}
	log.Printf("%d jeux chargés depuis l'API", len(cat.Games()))

	srv, err := server.New(cat, client, *web)
	if err != nil {
		log.Fatalf("templates introuvables (lancez le programme depuis le dossier gamedex ou utilisez -web) : %v", err)
	}

	addr := ":" + strconv.Itoa(*port)
	log.Printf("GameDex démarre sur http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, srv.Handler()))
}
