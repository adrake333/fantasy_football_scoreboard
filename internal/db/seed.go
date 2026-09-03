package db




import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/adrake333/fantasy_football_scoreboard/internal/config"
)




func SeedFromConfig(ctx context.Context, q Querier, cfg *config.Config) error {
	_, err := q.GetUserByUsername(ctx, "default_user")
	if err == nil {
		return nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("failed to check existing user: %w", err)
	}

	log.Println("Fresh database detected. Seeding users and leagues from config...")

	userID := "user_default"
	err = q.CreateUser(ctx, CreateUserParams{
		ID:				userID,
		Username:		"default_user",
		SleeperUserID:	sql.NullString{String: cfg.SleeperUserID, Valid: cfg.SleeperUserID != ""},
		EspnSwid:		sql.NullString{String: cfg.ESPNSWID, Valid: cfg.ESPNSWID != ""},
		EspnS2:			sql.NullString{String: cfg.ESPNS2, Valid: cfg.ESPNS2 != ""},
	})
	if err != nil {
		return fmt.Errorf("failed to seed default user: %w", err)
	}

	for i, l := range cfg.Leagues {
		leagueID := fmt.Sprintf("league_%d", i+1)
		err = q.CreateLeague(ctx, CreateLeagueParams{
			ID:					leagueID,
			UserID:				userID,
			Platform:			l.Platform,
			ExternalLeagueID:	l.ID,
			Name:				l.Alias,
			Season:				l.Season,
		})
		if err != nil {
			log.Printf("Warning: failed to see league %s: %v", l.Alias, err)
		}
	}

	log.Println("Database seeded successfully!")
	return nil
}