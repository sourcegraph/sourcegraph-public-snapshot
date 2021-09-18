BEGIN;

DROP INDEX IF EXISTS gitserver_repos_clone_status_idx;

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
