package fantasy




//import ()




type Platform string

const(
	PlatformSleeper	Platform = "sleeper"
	PlatformESPN	Platform = "espn"
)

type League struct {
	ID		string		`json:"id"`
	Name 		string 		`json:"name"`
	Platform	Platform	`json:"platform"`
	Season		string		`json:"season"`
}

type PlayerScore struct {
	PlayerID	string	`json:"player_id"`
	Name		string	`json:"name"`
	Position	string	`json:"position"`
	Points		float64	`json:"points"`
}

type RosterSpot struct {
	Slot	string		`json:"slot"`
	Player	PlayerScore	`json:"player"`
}

type Team struct {
	ID		string		`json:"id"`
	Name		string		`json:"name"`
	OwnerName	string		`json:"owner_name"`
	TotalScore	float64		`json:"total_score"`
	Roster		[]RosterSpot	`json:"roster"`
}

type Matchup struct {
	Leaguename	string
	UserTeam	string
	OpponentTeam	string
	UserScore	float64
	OpponentScore	float64
}
