BEGIN;

CREATE INDEX IF NOT EXISTS repo_metadata_gin_idx ON repo USING gin (metadata);

COMMIT;
