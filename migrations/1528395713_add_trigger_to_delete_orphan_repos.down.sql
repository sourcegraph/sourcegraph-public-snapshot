BEGIN;

DROP TRIGGER IF EXISTS trig_soft_delete_orphan_repos_for_external_service ON external_services;
DROP FUNCTION IF EXISTS soft_delete_orphan_repos();

COMMIT;
