BEGIN;

DROP TRIGGER IF EXISTS trig_delete_external_service_ref_on_external_service_repos
ON external_services;

DROP FUNCTION IF EXISTS delete_external_service_ref_on_external_service_repos();

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
