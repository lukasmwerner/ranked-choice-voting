package tallyballot

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/osu-acm/acm-votes/database"
	rankedchoice "github.com/osu-acm/acm-votes/ranked_choice"
)

func Main() {
	db, err := sql.Open("sqlite3", "data/data.db")
	if err != nil {
		log.Println("error opening database", err.Error())
		return
	}
	defer db.Close()
	queries := database.New(db)
	ctx := context.Background()

	id, err := uuid.Parse(os.Args[2])
	if err != nil {
		log.Println("ballot id not valid uuid", err.Error())
		return
	}
	winner, err := rankedchoice.RunBallot(db, id)
	if err != nil {
		log.Println("error running ranked choice", err.Error())
		return
	}

	ballotName, _ := queries.GetBallotName(ctx, id)
	name, err := queries.GetCandidateName(ctx, winner)
	if err != nil {
		log.Println("error getting winner name", err.Error(), winner)
		return
	}
	fmt.Printf("🎉 Winner[%s]: %s\n", ballotName, name)
}
