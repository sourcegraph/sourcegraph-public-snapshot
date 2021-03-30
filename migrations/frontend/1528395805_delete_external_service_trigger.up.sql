BEGIN;

DROP TRIGGER IF EXISTS trig_delete_external_service_ref_on_external_service_repos
ON external_services;

DROP FUNCTION IF EXISTS delete_external_service_ref_on_external_service_repos();

COMMIT;
