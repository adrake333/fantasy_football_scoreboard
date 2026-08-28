package main




import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/adrake333/fantasy_football_scoreboard/internal/api"
	"github.com/adrake333/fantasy_football_scoreboard/internal/fantasy"
)




type Config struct {
	ServerPort		string					`json:"server_port"`
	SleeperUserID	string					`json:"sleeper_user_id"`
	Leagues			[]fantasy.LeagueConfig	`json:"leagues"`
}

func main() {

	configData, err := os.ReadFile("config.dev.json") //CHANGE TO CONFIG.JSON WHEN LIVE
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	var cfg Config
	if err := json.Unmarshal(configData, &cfg); err != nil {
		log.Fatalf("Failed to parse config JSON: %v", err)
	}

	var sleeperLeagueIDs []string
	for _, league := range cfg.Leagues {
		if league.Platform == "sleeper" {
			sleeperLeagueIDs = append(sleeperLeagueIDs, league.ID)
		}
	}

	sleeperClient := fantasy.NewSleeperClient(10 * time.Second)

	server := &api.Server{
		SleeperClient:	sleeperClient,
		MyUserID:		cfg.SleeperUserID,
		Leagues:		cfg.Leagues,
	}

	http.HandleFunc("/api/matchups", server.HandleGetMatchups)
	http.HandleFunc("/", server.HandleDashboard)

	log.Printf("Server listening on %s...", cfg.ServerPort)
	log.Fatal(http.ListenAndServe(cfg.ServerPort, nil))
}
