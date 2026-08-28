package api




import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/adrake333/fantasy_football_scoreboard/internal/fantasy"
	"github.com/adrake333/fantasy_football_scoreboard/internal/web"
)




type Server struct {
	SleeperClient	*fantasy.SleeperClient
	LeagueIDs	[]string
}

func (s *Server) getMatchupsForWeek(week int) ([]fantasy.Matchup, error) {
	var allMatchups []fantasy.Matchup
	for _, leagueID := range s.LeagueIDs {
        	matchups, err := s.SleeperClient.FetchNormalizedMatchups(leagueID, week)
        	if err != nil {
            		return nil, err
        	}
        	allMatchups = append(allMatchups, matchups...)
    	}
    	return allMatchups, nil
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

func (s *Server) HandleGetMatchups(w http.ResponseWriter, r *http.Request) {
	week := s.getRequestedWeek(r)
	matchups, err := s.getMatchupsForWeek(week)
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
	matchups, err := s.getMatchupsForWeek(week)
	if err != nil {
		http.Error(w, "Failed to fetch matchups", http.StatusInternalServerError)
		return
	}

	availableWeeks := make([]int, 18)
	for i := 0; i < 18; i++ {
		availableWeeks[i] = i + 1
	}

	pageData := web.PageData{
		Week:		week,
		AvailableWeeks: availableWeeks,
		Matchups:	matchups,
	}

	err = web.RenderIndex(w, pageData)
	if err != nil {
		log.Printf("ERROR rendering template: %v", err)
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}
