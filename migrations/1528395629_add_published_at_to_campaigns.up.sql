BEGIN;

ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS published_at timestamptz;

UPDATE campaigns SET published_at = created_at WHERE published_at IS NULL;

COMMIT;
