// TP3 : GameDex appelle la mini API et affiche le résultat dans le terminal.
// Lancement (depuis le dossier gamedex) : go run ./cmd/tp3
package main

import (
	"flag"
	"fmt"
	"os"

	"gamedex/internal/api"
)

func main() {
	endpoint := flag.String("endpoint", "/api/games", "chemin de la liste des jeux (ex: /api/broken, /api/slow)")
	flag.Parse()

	client := api.New()
	client.GamesPath = *endpoint

	games, err := client.FetchGames()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Erreur :", err)
		os.Exit(1)
	}

	for _, g := range games {
		tournaments, err := client.FetchTournaments(g.ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erreur (tournois du jeu %d) : %v\n", g.ID, err)
			os.Exit(1)
		}
		fmt.Printf("%-25s %d  -  %d tournoi(s)\n", g.Name, g.Year, len(tournaments))
	}
}
