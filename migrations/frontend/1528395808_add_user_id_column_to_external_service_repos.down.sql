BEGIN;

DROP INDEX external_service_user_repos_idx;
ALTER TABLE external_service_repos DROP COLUMN user_id;

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
