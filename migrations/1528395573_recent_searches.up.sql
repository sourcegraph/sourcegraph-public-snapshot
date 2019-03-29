BEGIN;

CREATE TABLE recent_searches (
	id serial PRIMARY KEY,
	query text NOT NULL,
	created_at timestamp NOT NULL DEFAULT NOW()
);

COMMIT;
