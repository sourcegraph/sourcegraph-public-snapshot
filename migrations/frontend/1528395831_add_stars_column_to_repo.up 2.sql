BEGIN;

ALTER TABLE repo ADD COLUMN IF NOT EXISTS stars int;

UPDATE repo SET stars = (metadata->>'StargazerCount')::int
WHERE external_service_type = 'github'
AND stars IS NULL
AND metadata ? 'StargazerCount';

CREATE INDEX IF NOT EXISTS repo_stars_idx ON repo USING BTREE (stars DESC NULLS LAST);

COMMIT;
