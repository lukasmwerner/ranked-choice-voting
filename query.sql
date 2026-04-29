-- name: AddUser :exec
INSERT OR IGNORE INTO users (email, name) VALUES (?, ?);

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

-- name: GetCandidateName :one
SELECT name FROM candidates WHERE id = ?;

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

-- name: EliminateCandidateBallot :exec
UPDATE ballot_candidates
SET eliminated = TRUE
WHERE ballot_id = ? AND candidate_id = ?;

-- name: EliminateCandidateGlobally :exec
UPDATE ballot_candidates
SET eliminated = TRUE
WHERE candidate_id = ?;

-- name: CountRemaining :one
SELECT COUNT(*) FROM ballot_candidates
WHERE ballot_id = ? AND eliminated = FALSE;

-- VOTING

-- name: HasVoted :one
SELECT EXISTS (SELECT 1 FROM votes WHERE email = ? AND ballot_id = ?) AS has_voted;

-- name: DeleteVotes :exec
DELETE FROM votes WHERE email = ? AND ballot_id = ?;

-- name: InsertRankedVote :exec
INSERT INTO votes (email, ballot_id, candidate_id, rank)
VALUES (?, ?, ?, ?);

-- name: GetBallotStatus :one
SELECT status FROM ballots WHERE id = ?;

-- name: GetBallotName :one
SELECT name FROM ballots WHERE id = ?;

-- name: GetLastCandidate :one
SELECT candidate_id FROM ballot_candidates WHERE ballot_id = ? AND eliminated = FALSE LIMIT 1;

-- name: GetBallotCandidates :many
SELECT c.id, c.name
FROM candidates c
JOIN ballot_candidates bc ON bc.candidate_id = c.id
WHERE bc.ballot_id = ? AND bc.eliminated = FALSE
ORDER BY c.name;

-- name: GetBallotWinner :one
SELECT c.id, c.name
FROM candidates c
JOIN ballot_candidates bc ON bc.candidate_id = c.id
WHERE bc.ballot_id = ? AND bc.eliminated = FALSE
LIMIT 1;
