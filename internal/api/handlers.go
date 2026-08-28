package api




import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/adrake333/fantasy_football_scoreboard/internal/fantasy"
	"github.com/adrake333/fantasy_football_scoreboard/internal/web"
)




type Server struct {
	SleeperClient	*fantasy.SleeperClient
	MyUserID		string
	Leagues			[]fantasy.LeagueConfig
}

func (s *Server) getMatchupsForView(week int, leagueParam string) ([]fantasy.Matchup, error) {
	if leagueParam != "my_matchups" && leagueParam != "" {
		return s.SleeperClient.FetchNormalizedMatchups(leagueParam, week)
	}

	var myMatchups []fantasy.Matchup
	for _, league := range s.Leagues {
		matchups, err := s.SleeperClient.FetchNormalizedMatchups(league.ID, week)
		if err != nil {
			return nil, err
		}

		for _, m := range matchups {
			if m.UserOwnerID == s.MyUserID {
				myMatchups = append(myMatchups, m)
			} else if m.OpponentOwnerID == s.MyUserID {
				swapped := m
				swapped.UserOwnerID, swapped.OpponentOwnerID = m.OpponentOwnerID, m.UserOwnerID
				swapped.UserTeam, swapped.OpponentTeam = m.OpponentTeam, m.UserTeam
				swapped.UserScore, swapped.OpponentScore = m.OpponentScore, m.UserScore
				myMatchups = append(myMatchups, swapped)
			}
		}
	}
	return myMatchups, nil
}

func (s *Server) getRequestedWeek(r *http.Request) int {
	weekStr := r.URL.Query().Get("week")
	if weekStr != "" {
		if week, err := strconv.Atoi(weekStr); err == nil && week >= 1 && week <= 18 {
			return week
		}
	}

	currentWeek, err := s.SleeperClient.GetCurrentNFLWeek()
	if err != nil {
		return 1
	}

	return currentWeek
}

func (s *Server) getRequestedLeague(r *http.Request) string {
	league := r.URL.Query().Get("league")
	if league == "" {
		return "my_matchups"
	}
	return league
}

func (s *Server) HandleGetMatchups(w http.ResponseWriter, r *http.Request) {
	week := s.getRequestedWeek(r)
	leagueView := s.getRequestedLeague(r)

	matchups, err := s.getMatchupsForView(week, leagueView)
	if err != nil {
		http.Error(w, "Failed to fetch matchups", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(matchups)
}

func (s *Server) HandleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	week := s.getRequestedWeek(r)
	leagueView := s.getRequestedLeague(r)
	
	matchups, err := s.getMatchupsForView(week, leagueView)
	if err != nil {
		http.Error(w, "Failed to fetch matchups", http.StatusInternalServerError)
		return
	}

	availableWeeks := make([]int, 18)
	for i := 0; i < 18; i++ {
		availableWeeks[i] = i + 1
	}

	pageData := web.PageData{
		Week:			week,
		AvailableWeeks: availableWeeks,
		SelectedLeague: leagueView,
		Leagues:		s.Leagues,
		Matchups:		matchups,
	}

	err = web.RenderIndex(w, pageData)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}
