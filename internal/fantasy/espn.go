package fantasy




import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)




type ESPNClient struct{
	baseURL		string
	httpClient	*http.Client
	espnS2		string
	swid		string
}

func NewESPNClient(espnS2, swid string, timeout time.Duration) *ESPNClient {
	return &ESPNClient{
		baseURL:	"https://lm-api-reads.fantasy.espn.com/apis/v3/games/ffl/seasons",
		httpClient:	&http.Client{
			Timeout:	timeout,
		},
		espnS2:		espnS2,
		swid:		swid,
	}
}

type ESPNResponse struct {
	Settings struct {
		Name string `json:"name"`
	} `json:"settings"`
	Teams []ESPNTeam `json:"teams"`
	Schedule []ESPNMatchup `json:"schedule"`
}

type ESPNTeam struct {
    ID           int      `json:"id"`
    Location     string   `json:"location"`
    Nickname     string   `json:"nickname"`
    Name         string   `json:"name"`
    PrimaryOwner string   `json:"primaryOwner"`
    Owners       []string `json:"owners"`
}

func (t ESPNTeam) GetTeamName() string {
	if t.Name != "" {
		return t.Name
	}
	if t.Location != "" || t.Nickname != "" {
		return strings.TrimSpace(t.Location + " " + t.Nickname)
	}
	return fmt.Sprintf("Team %d", t.ID)
}

func (t ESPNTeam) GetOwnerID() string {
	if t.PrimaryOwner != "" {
		return t.PrimaryOwner
	}
	if len(t.Owners) > 0 {
		return t.Owners[0]
	}
	return ""
}

type ESPNMatchup struct {
    MatchupPeriodID int `json:"matchupPeriodId"`
    Home struct {
        TeamID     int     `json:"teamId"`
        TotalPoints float64 `json:"totalPoints"`
    } `json:"home"`
    Away struct {
        TeamID     int     `json:"teamId"`
        TotalPoints float64 `json:"totalPoints"`
    } `json:"away"`
}

func normalizeESPNData(data ESPNResponse, leagueID string, week int) []Matchup {
	leagueName := data.Settings.Name
	if leagueName == "" {
		leagueName = leagueID
	}

	teamMap := make(map[int]ESPNTeam)
	for _, team := range data.Teams {
		teamMap[team.ID] = team
	}

	var matchups []Matchup
	for _, m := range data.Schedule {
		if m.MatchupPeriodID != week {
			continue
		}

		if m.Home.TeamID == 0 || m.Away.TeamID == 0 {
			continue
		}

		homeTeam := teamMap[m.Home.TeamID]
		awayTeam := teamMap[m.Away.TeamID]

		matchup := Matchup{
			LeagueID:			leagueID,
			LeagueName:			leagueName,
			Week:				week,
			UserOwnerID:		homeTeam.GetOwnerID(),
			UserTeam:			homeTeam.GetTeamName(),
			UserScore:			m.Home.TotalPoints,
			OpponentOwnerID:	awayTeam.GetOwnerID(),
			OpponentTeam:		awayTeam.GetTeamName(),
			OpponentScore:		m.Away.TotalPoints,
		}

		matchups = append(matchups, matchup)
	}
	return matchups
}

func (c *ESPNClient) FetchNormalizedMatchups(leagueID, season string, week int) ([]Matchup, error) {
	if season == "" {
		season = "2026"
	}

	url := fmt.Sprintf("%s/%s/segments/0/leagues/%s?view=mMatchup&view=mTeam&view=mSettings&scoringPeriodID=%d", c.baseURL, season, leagueID, week)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if c.espnS2 != "" && c.swid != "" {
		cookieHeader := fmt.Sprintf("espn_s2=%s; SWID=%s", c.espnS2, c.swid)
		req.Header.Set("Cookie", cookieHeader)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("espn API error: status %d", resp.StatusCode)
	}

	var espnData ESPNResponse
	if err := json.NewDecoder(resp.Body).Decode(&espnData); err != nil {
		return nil, fmt.Errorf("failed to decode espn data: %w", err)
	}

	return normalizeESPNData(espnData, leagueID, week), nil
}