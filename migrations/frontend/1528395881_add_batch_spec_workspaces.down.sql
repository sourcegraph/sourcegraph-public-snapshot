BEGIN;

DROP TABLE IF EXISTS batch_spec_workspace_execution_jobs;
DROP TABLE IF EXISTS batch_spec_workspaces;
DROP TABLE IF EXISTS batch_spec_resolution_jobs;

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
