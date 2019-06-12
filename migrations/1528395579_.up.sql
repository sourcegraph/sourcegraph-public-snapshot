-- Adds an index on URI which is no queried more often. Not done in a
-- transaction so we can create it concurrently.

CREATE INDEX IF NOT EXISTS repo_uri_idx ON repo (uri);
