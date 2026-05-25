package initalizeelection

import (
	"context"
	"database/sql"
	"log"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"

	"github.com/osu-acm/acm-votes/database"
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

	// Candidates
	candidates := []struct {
		id   uuid.UUID
		name string
	}{
		{uuid.New(), "'Just For Fun' By Linus Torvalds & David Diamond"},                    // 0
		{uuid.New(), "'The Adventure of Tom Sawyer' by Mark Twain"},                         // 1
		{uuid.New(), "'A Year in Provionence' by Peter Male"},                               // 2
		{uuid.New(), "'The Idea Factory' by John Gertner"},                                  // 3
		{uuid.New(), "'Stoner' by John Williams"},                                           // 4
		{uuid.New(), "'The Count of Monte Cristo'"},                                         // 5
		{uuid.New(), "'Degrees of Freedom of Robotics and Social Justice' by Tom Williams"}, // 6
		{uuid.New(), "'How to Make a Hat Out of Dried Cucumber'"},                           // 7
		{uuid.New(), "'Educated' by Tara Westover"},                                         // 8
		{uuid.New(), "'A Psalm for the Wild Built' by Becky Chambers"},                      // 9
		{uuid.New(), "'Carless People' by Sarah Wynn-Williams"},                             // 10
		{uuid.New(), "'Is a River Alive?' by Robert Macfarlane"},                            // 11

	}

	for _, c := range candidates {
		if err := queries.AddCandidate(ctx, database.AddCandidateParams{
			ID:   c.id,
			Name: c.name,
		}); err != nil {
			log.Fatalf("add candidate %s: %v", c.name, err)
		}
	}

	// Ballots
	ballots := []struct {
		id          uuid.UUID
		name        string
		description string
		candidates  []uuid.UUID
	}{
		{
			id:          uuid.New(),
			name:        "Book Club Meeting",
			description: "",
			candidates: []uuid.UUID{
				candidates[0].id,
				candidates[1].id,
				candidates[2].id,
				candidates[3].id,
				candidates[4].id,
				candidates[5].id,
				candidates[6].id,
				candidates[7].id,
				candidates[8].id,
				candidates[9].id,
				candidates[10].id,
				candidates[11].id,
			},
		},
	}

	for _, b := range ballots {
		if err := queries.CreateBallot(ctx, database.CreateBallotParams{
			ID:          b.id,
			Name:        b.name,
			Description: b.description,
		}); err != nil {
			log.Fatalf("create ballot %s: %v", b.name, err)
		}

		for _, cid := range b.candidates {
			if err := queries.AddCandidateToBallot(ctx, database.AddCandidateToBallotParams{
				BallotID:    b.id,
				CandidateID: cid,
			}); err != nil {
				log.Fatalf("add candidate to ballot %s: %v", b.name, err)
			}
		}
	}

	log.Println("seeded successfully")
}
