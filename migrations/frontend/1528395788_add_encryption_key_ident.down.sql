BEGIN;

ALTER TABLE external_services DROP COLUMN IF EXISTS encryption_key_id;

COMMIT;
