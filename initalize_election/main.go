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
		{uuid.New(), "Lukas Werner"},           // 0
		{uuid.New(), "Avabella Schroeder"},     // 1
		{uuid.New(), "Trenston Ricks"},         // 2
		{uuid.New(), "Jacob Weiner"},           // 3
		{uuid.New(), "William Tu"},             // 4
		{uuid.New(), "Mehul Munankarmi"},       // 5
		{uuid.New(), "Leonidas Sallos"},        // 6
		{uuid.New(), "Cal (Caelynn) Corcoran"}, // 7
		{uuid.New(), "Isaac Tucknott"},         // 8
		{uuid.New(), "Krishna Basavaraju"},     // 9
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
			name:        "President",
			description: "",
			candidates:  []uuid.UUID{candidates[0].id, candidates[2].id},
		},
		{
			id:          uuid.New(),
			name:        "Vice-President",
			description: "",
			candidates:  []uuid.UUID{candidates[1].id},
		},
		{
			id:          uuid.New(),
			name:        "Treasurer",
			description: "",
			candidates:  []uuid.UUID{candidates[3].id, candidates[7].id, candidates[9].id},
		},
		{
			id:          uuid.New(),
			name:        "Head Competitive Officer",
			description: "",
			candidates:  []uuid.UUID{},
		},
		{
			id:          uuid.New(),
			name:        "Career Officer",
			description: "",
			candidates:  []uuid.UUID{candidates[2].id, candidates[5].id},
		},
		{
			id:          uuid.New(),
			name:        "Competitive Officer",
			description: "",
			candidates:  []uuid.UUID{candidates[4].id, candidates[8].id},
		},
		{
			id:          uuid.New(),
			name:        "Graphic Design Officer",
			description: "",
			candidates:  []uuid.UUID{candidates[7].id, candidates[9].id},
		},
		{
			id:          uuid.New(),
			name:        "Community Outreach Officer",
			description: "",
			candidates:  []uuid.UUID{candidates[1].id, candidates[3].id},
		},
		{
			id:          uuid.New(),
			name:        "Online Officer",
			description: "",
			candidates:  []uuid.UUID{candidates[6].id},
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
