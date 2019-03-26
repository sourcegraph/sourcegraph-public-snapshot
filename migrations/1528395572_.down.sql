BEGIN;
ALTER TABLE external_services DROP CONSTRAINT IF EXISTS check_non_empty_config;
COMMIT;
