BEGIN;

ALTER TABLE repo ADD COLUMN sources JSONB NOT NULL DEFAULT '{}'
CHECK (jsonb_typeof(sources) = 'object');
CREATE INDEX repo_sources_gin_idx ON repo USING GIN (sources);

COMMIT;
