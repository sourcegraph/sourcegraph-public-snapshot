BEGIN;

ALTER TABLE external_services DROP COLUMN IF EXISTS unrestricted;

COMMIT;
