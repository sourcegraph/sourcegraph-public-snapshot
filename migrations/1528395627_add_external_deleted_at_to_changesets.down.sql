BEGIN;

ALTER TABLE changesets DROP COLUMN IF EXISTS external_deleted_at;

COMMIT;
