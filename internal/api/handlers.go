package api




import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/adrake333/fantasy_football_scoreboard/internal/auth"
	"github.com/adrake333/fantasy_football_scoreboard/internal/db"
	"github.com/adrake333/fantasy_football_scoreboard/internal/fantasy"
	"github.com/adrake333/fantasy_football_scoreboard/internal/simulator"
	"github.com/adrake333/fantasy_football_scoreboard/internal/web"
	"github.com/google/uuid"
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

func (s *Server) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		session, err := s.DB.GetSession(r.Context(), cookie.Value)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		if session.ExpiresAt.Before(time.Now()) {
			_ = s.DB.DeleteSession(r.Context(), cookie.Value)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		ctx := context.WithValue(r.Context(), auth.UserContextKey, session.UserID)
		next(w, r.WithContext(ctx))
	}
}

func (s *Server) HandleGetMatchups(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := s.DB.GetUserByID(ctx, userID)
	if err != nil {
		http.Error(w, "Failed to load user profile", http.StatusInternalServerError)
		return
	}

	leagues, err := s.DB.GetLeaguesByUser(ctx, user.ID)
	if err != nil {
		http.Error(w, "Failed to load leagues from database", http.StatusInternalServerError)
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

	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := s.DB.GetUserByID(ctx, userID)
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

func (s *Server) HandleAddLeague(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	platform := r.Form.Get("platform")
	externalLeagueID := r.Form.Get("external_league_id")
	season := r.Form.Get("season")

	ctx := r.Context()

	userID, err := auth.GetUserID(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := s.DB.GetUserByID(ctx, userID)
	if err != nil {
		http.Error(w, "Failed to get user profile", http.StatusInternalServerError)
		return
	}

	leagueName := externalLeagueID

	if platform == "sleeper" {
		sleeperLeague, err := s.SleeperClient.GetLeague(externalLeagueID)
		if err != nil {
			http.Error(w, "Failed to fetch Sleeper league", http.StatusBadRequest)
			return
		}
		if sleeperLeague.Name != "" {
			leagueName = sleeperLeague.Name
		}
	}
	
	if platform == "espn" {
		espnLeague, err := s.ESPNClient.GetLeague(externalLeagueID, season)
		if err != nil {
			http.Error(w, "Failed to fetch ESPN league", http.StatusBadRequest)
			return
		}
		if espnLeague.Settings.Name != "" {
			leagueName = espnLeague.Settings.Name
		}
	}

	err = s.DB.CreateLeague(ctx, db.CreateLeagueParams{
		ID:					uuid.New().String(),
		UserID:				user.ID,
		Platform:			platform,
		ExternalLeagueID:	externalLeagueID,
		Name:				leagueName,
		Season:				season,
	})
	if err != nil {
		http.Error(w, "Failure adding league to database", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) HandleDeleteLeague(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	leagueID := r.Form.Get("league_id")

	err = s.DB.DeleteLeague(ctx, leagueID)
	if err != nil {
		http.Error(w, "Failure removing league from database", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}