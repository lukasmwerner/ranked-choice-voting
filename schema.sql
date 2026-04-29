CREATE TABLE users (
	email TEXT NOT NULL UNIQUE,
	name TEXT NOT NULL,

	PRIMARY KEY(email)
);

CREATE TABLE sessions (
	id TEXT primary key,
	user TEXT not null,
	expires_at INTEGER not null
);

CREATE TABLE candidates (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL
);

CREATE TABLE ballots (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending'
    -- 'pending' | 'open' | 'closed' | 'complete'
);

-- which candidates are on which ballot, and whether they've been eliminated
CREATE TABLE ballot_candidates (
    ballot_id UUID REFERENCES ballots(id),
    candidate_id UUID REFERENCES candidates(id),
    eliminated BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (ballot_id, candidate_id)
);

-- one row per ranking per voter per ballot
CREATE TABLE votes (
    email TEXT REFERENCES users(email) NOT NULL,
    ballot_id UUID REFERENCES ballots(id),
    candidate_id UUID REFERENCES candidates(id),
    rank INTEGER NOT NULL,  -- 1 = first choice, 2 = second, etc.
    PRIMARY KEY (email, ballot_id, rank)
);

