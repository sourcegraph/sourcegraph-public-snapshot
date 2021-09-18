BEGIN;

CREATE INDEX external_service_repos_external_service_id ON external_service_repos USING btree (external_service_id);

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
