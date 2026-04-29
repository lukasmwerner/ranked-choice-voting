package rankedchoice

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/osu-acm/acm-votes/database"
)

func RunBallot(db *sql.DB, ballotID uuid.UUID) (uuid.UUID, error) {
	queries := database.New(db)
	ctx := context.Background()

	status, err := queries.GetBallotStatus(ctx, ballotID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("RunBallot get status: %w", err)
	}
	if status != "closed" {
		return uuid.Nil, fmt.Errorf("RunBallot: ballot is not closed, current status: %s", status)
	}

	for {
		winner, err := database.CheckMajority(db, ballotID)
		if err != nil {
			return uuid.Nil, err
		}
		if winner != uuid.Nil {
			tx, err := db.Begin()
			if err != nil {
				return uuid.Nil, fmt.Errorf("RunBallot begin tx: %w", err)
			}
			qtx := queries.WithTx(tx)
			if err := qtx.CompleteBallot(ctx, ballotID); err != nil {
				tx.Rollback()
				return uuid.Nil, fmt.Errorf("RunBallot complete ballot: %w", err)
			}
			if err := qtx.EliminateCandidateGlobally(ctx, winner); err != nil {
				tx.Rollback()
				return uuid.Nil, fmt.Errorf("RunBallot global eliminate: %w", err)
			}
			if err := tx.Commit(); err != nil {
				return uuid.Nil, fmt.Errorf("RunBallot commit: %w", err)
			}
			return winner, nil
		}

		loser, isTie, err := database.FindLoser(db, ballotID)
		if err != nil {
			return uuid.Nil, err
		}
		if isTie {
			return uuid.Nil, fmt.Errorf("RunBallot: tie detected, manual resolution required")
		}

		if err := queries.EliminateCandidateBallot(ctx, database.EliminateCandidateBallotParams{
			BallotID:    ballotID,
			CandidateID: loser,
		}); err != nil {
			return uuid.Nil, fmt.Errorf("RunBallot eliminate loser: %w", err)
		}

		remaining, err := queries.CountRemaining(ctx, ballotID)
		if err != nil {
			return uuid.Nil, fmt.Errorf("RunBallot count remaining: %w", err)
		}
		if remaining == 0 {
			return uuid.Nil, fmt.Errorf("RunBallot: all candidates eliminated without a winner")
		}
		if remaining == 1 {
			winner, err := queries.GetLastCandidate(ctx, ballotID)
			if err != nil {
				return uuid.Nil, fmt.Errorf("RunBallot get last candidate: %w", err)
			}
			tx, err := db.Begin()
			if err != nil {
				return uuid.Nil, fmt.Errorf("RunBallot begin tx: %w", err)
			}
			qtx := queries.WithTx(tx)
			if err := qtx.CompleteBallot(ctx, ballotID); err != nil {
				tx.Rollback()
				return uuid.Nil, fmt.Errorf("RunBallot complete ballot: %w", err)
			}
			if err := qtx.EliminateCandidateGlobally(ctx, winner); err != nil {
				tx.Rollback()
				return uuid.Nil, fmt.Errorf("RunBallot global eliminate: %w", err)
			}
			if err := tx.Commit(); err != nil {
				return uuid.Nil, fmt.Errorf("RunBallot commit: %w", err)
			}
			return winner, nil
		}
	}
}
