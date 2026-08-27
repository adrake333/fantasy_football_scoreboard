package fantasy




import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)




type SleeperClient struct {
	baseURL		string
	httpClient	*http.Client
}

func NewSleeperClient(timeout time.Duration) *SleeperClient {
	return &SleeperClient{
		baseURL: "https://api.sleeper.app/v1",
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

type SleeperMatchup struct {
	RosterID	int			`json:"roster_id"`
	MatchupID	int			`json:"matchup_id"`
	Points		float64			`json:"points"`
	CustomPoints	float64			`json:"custom_points"`
	Starters	[]string		`json:"starters"`
	Players		[]string		`json:"players"`
	PlayersPoints	map[string]float64	`json:"players_points"`
}

func (c *SleeperClient) GetMatchups(leagueID string, week int) ([]SleeperMatchup, error) {
	url := fmt.Sprintf("%s/league/%s/matchups/%d", c.baseURL, leagueID, week)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("sleeper API error: status %d", resp.StatusCode)
	}

	var matchups []SleeperMatchup
	
	if err := json.NewDecoder(resp.Body).Decode(&matchups); err != nil {
		return nil, fmt.Errorf("failed to decode sleeper matchups: %w", err)
	}

	return matchups, nil
}

type SleeperRoster struct {
	RosterID	int	`json:"roster_id"`
	OwnerID		string	`json:"owner_id"`
}

func (c *SleeperClient) GetRosters(leagueID string) ([]SleeperRoster, error) {
	url := fmt.Sprintf("%s/league/%s/rosters", c.baseURL, leagueID)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("sleeper API error: status %d", resp.StatusCode)
	}

	var rosters []SleeperRoster

	if err := json.NewDecoder(resp.Body).Decode(&rosters); err != nil {
		return nil, fmt.Errorf("failed to decode sleeper rosters: %w", err)
	}

	return rosters, nil
}

type SleeperUser struct {
	UserID		string	`json:"user_id"`
	DisplayName	string	`json:"display_name"`
	Metadata	struct	{
		TeamName	string	`json:"team_name"`
	}	`json:"metadata"`
}

func (u SleeperUser) GetTeamName() string {
	if u.Metadata.TeamName != "" {
		return u.Metadata.TeamName
	}
	return u.DisplayName
}

func (c *SleeperClient) GetUsers(leagueID string) ([]SleeperUser, error) {
	url := fmt.Sprintf("%s/league/%s/users", c.baseURL, leagueID)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("sleeper API error: status %d", resp.StatusCode)
	}

	var users []SleeperUser

	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("failed to decode sleeper users: %w", err)
	}

	return users, nil
}
