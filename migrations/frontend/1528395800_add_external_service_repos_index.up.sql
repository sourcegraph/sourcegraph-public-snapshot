BEGIN;

CREATE INDEX external_service_repos_idx ON external_service_repos(external_service_id, repo_id);

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
