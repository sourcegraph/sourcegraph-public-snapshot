BEGIN;

ALTER TABLE external_services DROP COLUMN IF EXISTS last_sync_at;
ALTER TABLE external_services DROP COLUMN IF EXISTS next_sync_at;

COMMIT;
