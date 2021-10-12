BEGIN;

ALTER TABLE external_service_repos DROP COLUMN org_id;

COMMIT;
