BEGIN;

ALTER TABLE changeset_jobs DROP COLUMN IF EXISTS published_at;

COMMIT;
