-- +++
-- parent: 1528395848
-- +++

BEGIN;

-- Covered by external_service_repos_idx (external_service_id, repo_id)
DROP INDEX IF EXISTS external_service_repos_external_service_id;

COMMIT;
