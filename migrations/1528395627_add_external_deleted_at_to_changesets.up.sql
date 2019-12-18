BEGIN;

ALTER TABLE changesets ADD COLUMN external_deleted_at timestamptz;

COMMIT;
