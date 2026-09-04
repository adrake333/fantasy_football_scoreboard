package web




import (
	"html/template"
	"io"

	"github.com/adrake333/fantasy_football_scoreboard/internal/db"
	"github.com/adrake333/fantasy_football_scoreboard/internal/fantasy"
)




type PageData struct {
	Week			int
	AvailableWeeks	[]int
	SelectedLeague	string
	Leagues			[]db.League
	Matchups		[]fantasy.Matchup
}

func RenderIndex(w io.Writer, data PageData) error {
	tmpl, err := template.ParseFiles("internal/web/templates/index.html")
	if err != nil {
		return err
	}
	return tmpl.Execute(w, data)
}

func RenderRegister(w io.Writer) error {
	tmpl, err := template.ParseFiles("internal/web/templates/register.html")
	if err != nil {
		return err
	}
	return tmpl.Execute(w, nil)
}

func RenderLogin(w io.Writer) error {
	tmpl, err := template.ParseFiles("internal/web/templates/login.html")
	if err != nil {
		return err
	}
	return tmpl.Execute(w, nil)
}