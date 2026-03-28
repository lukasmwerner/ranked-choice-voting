-- name: AddUser :exec
INSERT OR IGNORE INTO users (
	email, first_name, last_name
) VALUES (
	?, ?, ?
);

-- name: UserExists :one
SELECT 1 FROM users
WHERE email = ?;

-- name: CreateBallot :exec
INSERT INTO ballots (
	id, name, description, options
) VALUES (
	?, ?, ?, JSONB(?)
);

-- name: GetBallotVotes :many
SELECT * FROM votes
WHERE ballot = ?;


-- name: SubmitVote :one
INSERT INTO votes (
	id, email,
	ballot, vote
) VALUES (
	?, ?, ?, ?
) RETURNING *;

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
