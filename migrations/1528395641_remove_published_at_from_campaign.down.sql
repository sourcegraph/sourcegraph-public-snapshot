BEGIN;

ALTER TABLE campaigns ADD COLUMN published_at timestamptz;

COMMIT;
