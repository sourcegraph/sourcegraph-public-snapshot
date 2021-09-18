BEGIN;

DROP INDEX IF EXISTS external_service_repos_idx;

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
