BEGIN;

ALTER TABLE changesets ADD COLUMN IF NOT EXISTS external_updated_at timestamptz;

-- Safe enough default, the true value will be computed at next sync
UPDATE changesets
SET external_updated_at = updated_at;

COMMIT;
