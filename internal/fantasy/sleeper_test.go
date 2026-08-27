package fantasy




import "testing"




func TestNormalizeMatchups(t *testing.T) {
	mockUsers := []SleeperUser{
		{UserID: "101", DisplayName: "Adam", Metadata: struct{ TeamName string `json:"team_name"` }{TeamName: "Moby Ditka"}},
		{UserID: "202", DisplayName: "Allison", Metadata: struct{ TeamName string `json:"team_name"` }{TeamName: "Baby Got Dak"}},
	}
	mockRosters := []SleeperRoster{
		{RosterID: 1, OwnerID: "101"},
		{RosterID: 2, OwnerID: "202"},
	}
	mockMatchups := []SleeperMatchup{
		{MatchupID: 1, RosterID: 1, Points: 105.4},
		{MatchupID: 1, RosterID: 2, Points: 98.2},
	}

	result := normalizeData(mockMatchups, mockRosters, mockUsers, "test-league", 1)

	if len(result) != 1 {
		t.Fatalf("expected 1 matchup, got %d", len(result))
	}

	m := result[0]
	if m.UserScore != 105.4 {
		t.Errorf("expected score 105.4, got %.2f", m.UserScore)
	}
}
