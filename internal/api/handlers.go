package api




import (
	"log"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/adrake333/fantasy_football_scoreboard/internal/db"
	"github.com/adrake333/fantasy_football_scoreboard/internal/fantasy"
	"github.com/adrake333/fantasy_football_scoreboard/internal/simulator"
	"github.com/adrake333/fantasy_football_scoreboard/internal/web"
)




type Server struct {
	DB				db.Querier
	SleeperClient	*fantasy.SleeperClient
	ESPNClient		*fantasy.ESPNClient
	Simulator		*simulator.Simulator
}

func matchesUserID(ownerID, targetID string) bool {
	if ownerID == "" {
		return false
	}

	cleanOwner := strings.ToUpper(strings.Trim(ownerID, "{}"))
	cleanTargetID := strings.ToUpper(strings.Trim(targetID, "{}"))
	return cleanTargetID != "" && cleanOwner == cleanTargetID
}

func (s *Server) getMatchupsForView(ctx context.Context, week int, leagueParam string, user *db.User, leagues []db.League) ([]fantasy.Matchup, error) {
	if leagueParam != "my_matchups" && leagueParam != "" {
		for _, l := range leagues {
			if l.ID == leagueParam {
				if l.Platform == "espn" {
					return s.ESPNClient.FetchNormalizedMatchups(l.ExternalLeagueID, l.Season, week)
				}
				return s.SleeperClient.FetchNormalizedMatchups(l.ExternalLeagueID, week)
			}
		}
	}

	var myMatchups []fantasy.Matchup
	for _, league := range leagues {
		var matchups []fantasy.Matchup
		var err error
		
		if league.Platform == "espn" {
			matchups, err = s.ESPNClient.FetchNormalizedMatchups(league.ExternalLeagueID, league.Season, week)
		} else {
			matchups, err = s.SleeperClient.FetchNormalizedMatchups(league.ExternalLeagueID, week)
		}

		if err != nil {
			return nil, err
		}

		targetID := user.SleeperUserID.String
		if league.Platform == "espn" {
			targetID = user.EspnSwid.String
		}

		for _, m := range matchups {

			isUserTeam := matchesUserID(m.UserOwnerID, targetID)
			isOpponent := matchesUserID(m.OpponentOwnerID, targetID)
			
			if isUserTeam {
				myMatchups = append(myMatchups, m)
			} else if isOpponent {
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
	ctx := r.Context()

	user, err := s.DB.GetUserByUsername(ctx, "default_user")
	if err != nil {
		http.Error(w, "Failed to load user profile", http.StatusInternalServerError)
		return
	}

	leagues, err := s.DB.GetLeaguesByUser(ctx, user.ID)
	if err != nil {
		http.Error(w, "Failed to laod leagues from database", http.StatusInternalServerError)
		return
	}
	
	week := s.getRequestedWeek(r)
	leagueView := s.getRequestedLeague(r)

	matchups, err := s.getMatchupsForView(ctx, week, leagueView, &user, leagues)
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

	ctx := r.Context()

	user, err := s.DB.GetUserByUsername(ctx, "default_user")
	if err != nil {
		http.Error(w, "Failed to load user profile", http.StatusInternalServerError)
		return
	}

	leagues, err := s.DB.GetLeaguesByUser(ctx, user.ID)
	if err != nil {
		http.Error(w, "Failed to laod leagues from database", http.StatusInternalServerError)
		return
	}

	week := s.getRequestedWeek(r)
	leagueView := s.getRequestedLeague(r)
	
	matchups, err := s.getMatchupsForView(ctx, week, leagueView, &user, leagues)
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
		Leagues:		leagues,
		Matchups:		matchups,
	}

	err = web.RenderIndex(w, pageData)
	if err != nil {
//debug log
		log.Printf("[TEMPLATE ERROR]: %v", err)
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}

func (s *Server) HandleStream(w http.ResponseWriter, r *http.Request) {
	if s.Simulator == nil {
		http.Error(w, "Simulation is not active", http.StatusServiceUnavailable)
		return
	}
	
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	msgChan := s.Simulator.Subscribe()
	defer s.Simulator.Unsubscribe(msgChan)

	for {
		select {
		case <-r.Context().Done():
			return
		case matchups, ok := <-msgChan:
			if !ok {
				return
			}

			data, err := json.Marshal(matchups)
			if err != nil {
				continue
			}

			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}