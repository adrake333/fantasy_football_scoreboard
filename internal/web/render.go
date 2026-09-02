package web




import (
	"html/template"
	"io"

	"github.com/adrake333/fantasy_football_scoreboard/internal/config"
	"github.com/adrake333/fantasy_football_scoreboard/internal/fantasy"
)




type PageData struct {
	Week			int
	AvailableWeeks	[]int
	SelectedLeague	string
	Leagues			[]config.LeagueConfig
	Matchups		[]fantasy.Matchup
}

func RenderIndex(w io.Writer, data PageData) error {
	tmpl, err := template.ParseFiles("internal/web/templates/index.html")
	if err != nil {
		return err
	}
	return tmpl.Execute(w, data)
}
