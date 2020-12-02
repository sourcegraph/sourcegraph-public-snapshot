BEGIN;

DROP TRIGGER IF EXISTS trig_soft_delete_orphan_repo_by_external_service_repo ON external_service_repos;
DROP FUNCTION IF EXISTS soft_delete_orphan_repo_by_external_service_repos() CASCADE;

COMMIT;
