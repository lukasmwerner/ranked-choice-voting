-- name: AddUser :exec
INSERT OR IGNORE INTO users (
	email, first_name, last_name
) VALUES (
	?, ?, ?
);

-- name: UserExists :one
SELECT 1 FROM users
WHERE email = ?;

-- name: GetSession :one
SELECT user
FROM sessions
WHERE id = ? AND
	  expires_at > strftime('%s', 'now');

-- name: CreateSession :exec
INSERT INTO sessions (
	id, user, expires_at
) VALUES (?, ?, ?);

-- name: CleanSessions :exec
DELETE FROM sessions WHERE
expires_at < strftime('%s', 'now');

-- BALLOT SETUP

-- name: AddCandidate :exec
INSERT INTO candidates (id, name) VALUES (?, ?);

-- name: CreateBallot :exec
INSERT INTO ballots (id, name, description) VALUES (?, ?, ?);

-- name: AddCandidateToBallot :exec
INSERT INTO ballot_candidates (ballot_id, candidate_id) VALUES (?, ?);

-- name: OpenBallot :exec
UPDATE ballots SET status = 'open' WHERE id = ?;

-- name: CloseBallot :exec
UPDATE ballots SET status = 'closed' WHERE id = ?;

-- name: CompleteBallot :exec
UPDATE ballots SET status = 'complete' WHERE id = ?;

-- name: EliminateCandidate :exec
UPDATE ballot_candidates SET eliminated = TRUE WHERE candidate_id = ?;

-- name: CheckMajority :many
WITH active_first_choice AS (
    SELECT v.email, v.candidate_id, v.rank
    FROM votes v
    JOIN ballot_candidates bc
        ON bc.ballot_id = v.ballot_id
        AND bc.candidate_id = v.candidate_id
    WHERE v.ballot_id = ?
      AND bc.eliminated = FALSE
      AND v.rank = (
          -- find this voter's current effective first choice
          SELECT MIN(v2.rank)
          FROM votes v2
          JOIN ballot_candidates bc2
              ON bc2.ballot_id = v2.ballot_id
              AND bc2.candidate_id = v2.candidate_id
          WHERE v2.email = v.email
            AND v2.ballot_id = v.ballot_id
            AND bc2.eliminated = FALSE
      )
)
SELECT candidate_id, COUNT(*) as vote_count
FROM active_first_choice
GROUP BY candidate_id
ORDER BY vote_count DESC;

--- name FindLoser :many

