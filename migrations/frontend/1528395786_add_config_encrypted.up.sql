BEGIN;

ALTER TABLE external_services ADD COLUMN IF NOT EXISTS config_encrypted boolean;

COMMIT;
