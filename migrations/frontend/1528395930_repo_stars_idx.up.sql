BEGIN;

DROP INDEX IF EXISTS repo_stars_idx;
CREATE INDEX IF NOT EXISTS repo_stars_idx ON repo (stars DESC NULLS LAST) INCLUDE (id, name, private) WHERE (deleted_at IS NULL AND blocked IS NULL);

COMMIT;
