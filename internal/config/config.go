package config




import(
	"encoding/json"
	"fmt"
	"os"
)




type Config struct {
	ServerPort		string					`json:"server_port"`
	SleeperUserID	string					`json:"sleeper_user_id"`
	ESPNS2			string					`json:"espn_s2"`
	ESPNSWID		string					`json:"espn_swid"`
	Leagues			[]LeagueConfig			`json:"leagues"`
}

type LeagueConfig struct {
	ID			string	`json:"id"`
	Platform	string	`json:"platform"`
	Alias		string	`json:"alias"`
	Season		string	`json:"season"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("could not parse config JSON: %w", err)
	}

	return &cfg, nil
}