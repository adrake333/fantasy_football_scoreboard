package main




import (
//	"flag"
	"fmt"
	"time"

	"github.com/adrake333/fantasy_football_scoreboard/internal/fantasy"
)




func main() {
//	configPath := flag.String("config", "config.json", "path to configuration file")
//	flag.Parse()

	sleeperClient := fantasy.NewSleeperClient(10 * time.Second)

	matchups, err := sleeperClient.GetMatchups("1180280607007068160", 1)
	if err != nil {
		fmt.Println("unable to get sleeper matchups: %s", err)
		return
	}

	rosters, err := sleeperClient.GetRosters("1180280607007068160")
	if err != nil {
		fmt.Println("unable to get sleeper rosters: %s", err)
		return
	}

	users, err := sleeperClient.GetUsers("1180280607007068160")
	if err != nil {
		fmt.Println("unable to get sleeper users: %s", err)
		return
	}

	fmt.Printf("%+v\n%+v\n%+v\n", matchups, rosters, users)
}
