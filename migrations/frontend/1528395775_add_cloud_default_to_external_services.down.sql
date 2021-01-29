BEGIN;

ALTER TABLE external_services DROP COLUMN IF EXISTS cloud_default;
DROP INDEX IF EXISTS kind_cloud_default;

COMMIT;
