BEGIN;

DROP TRIGGER IF EXISTS trig_delete_repo_ref_on_external_service_repos ON repo;
DROP TRIGGER IF EXISTS trig_delete_external_service_ref_on_external_service_repos ON external_services;

DROP FUNCTION IF EXISTS delete_repo_ref_on_external_service_repos();
DROP FUNCTION IF EXISTS delete_external_service_ref_on_external_service_repos();

DROP TABLE IF EXISTS external_service_repos;
ALTER TABLE repo ADD COLUMN sources jsonb DEFAULT '{}'::jsonb NOT NULL;

COMMIT;
