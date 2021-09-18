BEGIN;

-- Covered by external_service_repos_idx (external_service_id, repo_id)
DROP INDEX IF EXISTS external_service_repos_external_service_id;

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
