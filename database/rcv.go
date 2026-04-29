package database

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

const activeFirstChoiceCTE = `
WITH active_first_choice AS (
    SELECT v.email, v.candidate_id
    FROM votes v
    JOIN ballot_candidates bc
        ON bc.ballot_id = v.ballot_id
        AND bc.candidate_id = v.candidate_id
    WHERE v.ballot_id = ?
      AND bc.eliminated = FALSE
      AND v.rank = (
          SELECT MIN(v2.rank)
          FROM votes v2
          JOIN ballot_candidates bc2
              ON bc2.ballot_id = v2.ballot_id
              AND bc2.candidate_id = v2.candidate_id
          WHERE v2.email = v.email
            AND v2.ballot_id = v.ballot_id
            AND bc2.eliminated = FALSE
      )
)`

func CheckMajority(db *sql.DB, ballotID uuid.UUID) (uuid.UUID, error) {
	// returns candidate_id if someone has >50%, else ""
	query := activeFirstChoiceCTE + `
    , totals AS (
        SELECT candidate_id, COUNT(*) as vote_count
        FROM active_first_choice
        GROUP BY candidate_id
    ),
    total_votes AS (
        SELECT SUM(vote_count) as total FROM totals
    )
    SELECT t.candidate_id
    FROM totals t, total_votes tv
    WHERE CAST(t.vote_count AS FLOAT) / tv.total > 0.5
    LIMIT 1;`

	var winnerID string
	err := db.QueryRow(query, ballotID).Scan(&winnerID)
	if err == sql.ErrNoRows {
		return uuid.Nil, nil
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("checkMajority: %w", err)
	}
	uid, _ := uuid.Parse(winnerID)
	return uid, nil
}

func FindLoser(db *sql.DB, ballotID uuid.UUID) (id uuid.UUID, isTie bool, err error) {
	// returns the candidate with fewest votes, and whether there's a tie
	query := activeFirstChoiceCTE + `
    SELECT candidate_id, COUNT(*) as vote_count
    FROM active_first_choice
    GROUP BY candidate_id
    ORDER BY vote_count ASC
    LIMIT 2;`

	rows, err := db.Query(query, ballotID)
	if err != nil {
		return uuid.Nil, false, fmt.Errorf("findLoser: %w", err)
	}
	defer rows.Close()

	type result struct {
		id    string
		votes int
	}
	var results []result

	for rows.Next() {
		var r result
		if err := rows.Scan(&r.id, &r.votes); err != nil {
			return uuid.Nil, false, fmt.Errorf("findLoser scan: %w", err)
		}
		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return uuid.Nil, false, fmt.Errorf("findLoser rows: %w", err)
	}
	if len(results) == 0 {
		return uuid.Nil, false, fmt.Errorf("findLoser: no candidates remaining")
	}
	if len(results) >= 2 && results[0].votes == results[1].votes {
		return uuid.Nil, true, nil
	}
	uid, _ := uuid.Parse(results[0].id)
	return uid, false, nil
}

func MarkBallotComplete(db *sql.DB, tx *sql.Tx, ballotID, winnerID uuid.UUID) error {
	_, err := tx.Exec(`
        UPDATE ballots SET status = 'complete' WHERE id = ?
    `, ballotID)
	if err != nil {
		return fmt.Errorf("markBallotComplete: %w", err)
	}
	_, err = tx.Exec(`
        UPDATE ballot_candidates
        SET eliminated = TRUE
        WHERE candidate_id = ? AND ballot_id != ?
    `, winnerID, ballotID)
	if err != nil {
		return fmt.Errorf("eliminateGlobally: %w", err)
	}
	return nil
}
