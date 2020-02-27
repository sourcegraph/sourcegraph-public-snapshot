BEGIN;

ALTER TABLE changesets DROP COLUMN IF EXISTS external_updated_at;

COMMIT;
