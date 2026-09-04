package api




import (
	"database/sql"
	"net/http"
	"time"

	"github.com/adrake333/fantasy_football_scoreboard/internal/auth"
	"github.com/adrake333/fantasy_football_scoreboard/internal/db"
	"github.com/adrake333/fantasy_football_scoreboard/internal/web"
	"github.com/google/uuid"
)




func (s *Server) HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		if err := web.RenderRegister(w); err != nil {
			http.Error(w, "Failed to render register page", http.StatusInternalServerError)
		}
		return
	}
	
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	username := r.Form.Get("username")
	password := r.Form.Get("password")
	sleeperUserID := r.Form.Get("sleeper_user_id")
	espnSWID := r.Form.Get("espn_swid")
	espnS2 := r.Form.Get("espn_s2")

	if username == "" || password == "" {
		http.Error(w, "Please enter a username and password", http.StatusBadRequest)
		return
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		http.Error(w, "Failed to secure password", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	userID := uuid.New().String()

	err = s.DB.CreateUser(ctx, db.CreateUserParams{
		ID:				userID,
		Username:		username,
		PasswordHash:	hash,
		SleeperUserID:	sql.NullString{String: sleeperUserID, Valid: sleeperUserID != ""},
		EspnSwid:		sql.NullString{String: espnSWID, Valid: espnSWID != ""},
		EspnS2:			sql.NullString{String: espnS2, Valid: espnS2 != ""},
	})
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	token, err := auth.GenerateSessionToken()
	if err != nil {
		http.Error(w, "Failed to create session token", http.StatusInternalServerError)
		return
	}

	expiresAt := time.Now().Add(24 * 7 * time.Hour)

	err = s.DB.CreateSession(ctx, db.CreateSessionParams{
		Token:		token,
		UserID:		userID,
		ExpiresAt:	expiresAt,
	})
	if err != nil {
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:		"session_token",
		Value:		token,
		Expires:	expiresAt,
		HttpOnly:	true,
		Path:		"/",
		SameSite:	http.SameSiteLaxMode,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		if err := web.RenderLogin(w); err != nil {
			http.Error(w, "Failed to render login page", http.StatusInternalServerError)
		}
		return
	}
	
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	username := r.Form.Get("username")
	password := r.Form.Get("password")
	ctx := r.Context()

	user, err := s.DB.GetUserByUsername(ctx, username)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusBadRequest)
		return
	}

	if !auth.CheckPasswordHash(password, user.PasswordHash) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized);
		return
	}

	token, err := auth.GenerateSessionToken()
	if err != nil {
		http.Error(w, "Failed to create session token", http.StatusInternalServerError)
		return
	}

	expiresAt := time.Now().Add(24 * 7 * time.Hour)

	err = s.DB.CreateSession(ctx, db.CreateSessionParams{
		Token:		token,
		UserID:		user.ID,
		ExpiresAt:	expiresAt,
	})
	if err != nil {
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
    	Name:     "session_token",
    	Value:    token,
    	Expires:  expiresAt,
    	HttpOnly: true,
    	Path:     "/",
    	SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) HandleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if cookie, err := r.Cookie("session_token"); err == nil {
		_ = s.DB.DeleteSession(r.Context(), cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:		"session_token",
		Value:		"",
		MaxAge:		-1,
		Expires:	time.Unix(0, 0),
		Path:		"/",
		HttpOnly:	true,
		SameSite:	http.SameSiteLaxMode,
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}