BEGIN;

ALTER TABLE external_services DROP COLUMN IF EXISTS config_encrypted;

COMMIT;
