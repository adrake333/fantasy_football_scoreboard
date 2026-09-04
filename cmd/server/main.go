package main




import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/adrake333/fantasy_football_scoreboard/internal/api"
	"github.com/adrake333/fantasy_football_scoreboard/internal/config"
	"github.com/adrake333/fantasy_football_scoreboard/internal/db"
	"github.com/adrake333/fantasy_football_scoreboard/internal/fantasy"
	"github.com/adrake333/fantasy_football_scoreboard/internal/simulator"
)




func main() {
	simulateFlag := flag.Bool("simulate", false, "Enable live simulation mode for streaming scores")
	configPath := flag.String("config", "config.json", "Path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	dbConn, err := db.Init("scoreboard.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer dbConn.Close()

	dbQueries := db.New(dbConn)

//	if err := db.SeedFromConfig(context.Background(), dbQueries, cfg); err != nil {
//		log.Printf("Seeding warning: %v", err)
//	}

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
		DB:				dbQueries,
		SleeperClient:	sleeperClient,
		ESPNClient:		espnClient,
		Simulator:		sim,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/login", server.HandleLogin)
    mux.HandleFunc("/register", server.HandleRegister)

	mux.HandleFunc("/", server.RequireAuth(server.HandleDashboard))
    mux.HandleFunc("/logout", server.RequireAuth(server.HandleLogout))
    mux.HandleFunc("/api/matchups", server.RequireAuth(server.HandleGetMatchups))
    mux.HandleFunc("/api/stream", server.RequireAuth(server.HandleStream))
    mux.HandleFunc("/leagues/add", server.RequireAuth(server.HandleAddLeague))
    mux.HandleFunc("/leagues/delete", server.RequireAuth(server.HandleDeleteLeague))

	log.Printf("Server listening on %s...", cfg.ServerPort)
	log.Fatal(http.ListenAndServe(cfg.ServerPort, mux))
}