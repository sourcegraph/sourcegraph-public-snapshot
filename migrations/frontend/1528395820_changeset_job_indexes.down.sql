BEGIN;

DROP INDEX IF EXISTS changeset_jobs_state_idx;
DROP INDEX IF EXISTS changeset_jobs_bulk_group_idx;

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
