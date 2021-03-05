BEGIN;

ALTER TABLE external_service_repos
DROP CONSTRAINT IF EXISTS external_service_repos_repo_id_external_service_id_unique;

COMMIT;
