BEGIN;

ALTER TABLE campaigns ADD COLUMN published_at timestamptz;

UPDATE campaigns SET published_at = created_at;

COMMIT;
