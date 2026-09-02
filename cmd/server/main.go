package main




import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/adrake333/fantasy_football_scoreboard/internal/api"
	"github.com/adrake333/fantasy_football_scoreboard/internal/config"
	"github.com/adrake333/fantasy_football_scoreboard/internal/fantasy"
	"github.com/adrake333/fantasy_football_scoreboard/internal/simulator"
)




func main() {
	simulateFlag := flag.Bool("simulate", false, "Enable live simulation mode for streaming scores")
	configPath := flag.String("config", "config.dev.json", "Path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	sleeperClient := fantasy.NewSleeperClient(10 * time.Second)
	espnClient := fantasy.NewESPNClient(cfg.ESPNS2, cfg.ESPNSWID, 10 * time.Second)

	var sim *simulator.Simulator
	if *simulateFlag {
		log.Println("Seeding and starting score simulator in background...")
		var initialMatchups []fantasy.Matchup
		for _, league := range cfg.Leagues {
			var matchups []fantasy.Matchup
			var err error

			if league.Platform == "sleeper" {
				matchups, err = sleeperClient.FetchNormalizedMatchups(league.ID, 1)
			} else if league.Platform == "espn" {
				matchups, err = espnClient.FetchNormalizedMatchups(league.ID, league.Season, 1)
			}

			if err != nil {
				log.Printf("Warning: could not seed simulator for %s league %s: %v", league.Platform, league.ID, err)
				continue
			}
			
			initialMatchups = append(initialMatchups, matchups...)
		}

		sim = simulator.NewSimulator(initialMatchups)
		sim.Start(3 * time.Second)
	}

	server := &api.Server{
		SleeperClient:	sleeperClient,
		ESPNClient:		espnClient,
		MyUserID:		cfg.SleeperUserID,
		ESPNSWID:		cfg.ESPNSWID,
		Leagues:		cfg.Leagues,
		Simulator:		sim,
	}

	http.HandleFunc("/api/matchups", server.HandleGetMatchups)
	http.HandleFunc("/", server.HandleDashboard)
	http.HandleFunc("/api/stream", server.HandleStream)

	log.Printf("Server listening on %s...", cfg.ServerPort)
	log.Fatal(http.ListenAndServe(cfg.ServerPort, nil))
}