BEGIN;

ALTER TABLE changeset_jobs ADD COLUMN published_at timestamptz;

COMMIT;
