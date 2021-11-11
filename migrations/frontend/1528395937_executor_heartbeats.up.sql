BEGIN;

CREATE TABLE IF NOT EXISTS executor_heartbeats (
    id SERIAL PRIMARY KEY,
    hostname TEXT NOT NULL UNIQUE,
    -- TODO - add _cool_ fields
    last_seen_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- TODO - add comments

COMMIT;
