BEGIN;

ALTER TABLE external_service_repos DROP COLUMN IF EXISTS org_id;

COMMIT;
