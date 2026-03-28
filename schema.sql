CREATE TABLE users (
	email TEXT NOT NULL UNIQUE,
	first_name TEXT NOT NULL,
	last_name TEXT NOT NULL,

	PRIMARY KEY(email)
);

CREATE TABLE votes (
	id uuid UNIQUE, -- idempotent id
	email text not null, -- user email
	ballot uuid, -- ballot id
	vote text not null,

	PRIMARY KEY(email, ballot)
);

CREATE TABLE ballots (
	id uuid UNIQUE,
	name text not null,
	description text not null,
	options blob
);

CREATE TABLE sessions (
	id TEXT primary key,
	user TEXT not null,
	expires_at INTEGER not null
);
