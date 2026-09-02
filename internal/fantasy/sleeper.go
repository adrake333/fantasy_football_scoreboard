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

func normalizeData(matchups []SleeperMatchup, rosters []SleeperRoster, users []SleeperUser, leagueID string, week int) []Matchup {
	userMap := make(map[string]SleeperUser)
	for _, u := range users {
		userMap[u.UserID] = u
	}

	rosterOwnerMap := make(map[int]string)
	for _, r := range rosters {
		rosterOwnerMap[r.RosterID] = r.OwnerID
	}

	matchupsByID := make(map[int][]SleeperMatchup)
	for _, m := range matchups {
		if m.MatchupID == 0 {
			continue
		}
		matchupsByID[m.MatchupID] = append(matchupsByID[m.MatchupID], m)
	}

	var normalized []Matchup

	for _, teamPair := range matchupsByID {
		if len(teamPair) < 2 {
			continue
		}

		rawTeamA := teamPair[0]
		rawTeamB := teamPair[1]

		ownerIDA := rosterOwnerMap[rawTeamA.RosterID]
		userA := userMap[ownerIDA]
		teamA := Team{
			ID:		ownerIDA,
			Name:		userA.GetTeamName(),
			OwnerName:	userA.DisplayName,
			TotalScore:	rawTeamA.Points,
		}

		ownerIDB := rosterOwnerMap[rawTeamB.RosterID]
		userB := userMap[ownerIDB]
		teamB := Team{
			ID:		ownerIDB,
			Name:		userB.GetTeamName(),
			OwnerName:	userB.DisplayName,
			TotalScore:	rawTeamB.Points,
		}

		matchup := Matchup{
			LeagueID:			leagueID,
			LeagueName:			leagueID,
			Week:				week,
			UserOwnerID:		ownerIDA,
			UserTeam:			teamA.Name,
			UserScore:			teamA.TotalScore,
			OpponentOwnerID:	ownerIDB,
			OpponentTeam:		teamB.Name,
			OpponentScore:		teamB.TotalScore,
		}

		normalized = append(normalized, matchup)
	}

	return normalized
}


func (c *SleeperClient) FetchNormalizedMatchups(leagueID string, week int) ([]Matchup, error) {

	matchups, err := c.GetMatchups(leagueID, week)
	if err != nil {
		return nil, err
	}

	rosters, err := c.GetRosters(leagueID)
	if err != nil {
		return nil, err
	}

	users, err := c.GetUsers(leagueID)
	if err != nil {
		return nil, err
	}

	return normalizeData(matchups, rosters, users, leagueID, week), nil
}

type NFLState struct {
	Week		int	`json:"week"`
	Season		string	`json:"season"`
	SeasonType	string	`json:"season_type"`
}

func (c *SleeperClient) GetCurrentNFLWeek() (int, error) {
	url := fmt.Sprintf("%s/state/nfl", c.baseURL)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return 1, err
	}
	defer resp.Body.Close()

	var state NFLState
	if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
		return 1, err
	}

	if state.Week < 1 {
		return 1, nil
	}

	return state.Week, nil
}
