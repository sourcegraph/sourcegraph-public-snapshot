BEGIN;

DROP INDEX repo_sources_gin_idx;
ALTER TABLE repo DROP COLUMN sources;

COMMIT;
