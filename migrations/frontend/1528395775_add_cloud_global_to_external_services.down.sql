BEGIN;

ALTER TABLE external_services DROP COLUMN IF EXISTS cloud_global;
DROP INDEX IF EXISTS kind_cloud_global;

COMMIT;
