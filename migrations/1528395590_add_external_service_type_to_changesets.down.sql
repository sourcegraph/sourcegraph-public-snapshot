BEGIN;

ALTER TABLE changesets DROP COLUMN IF EXISTS external_service_type;

COMMIT;
