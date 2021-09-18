BEGIN;
DROP INDEX IF EXISTS lsif_dependency_indexing_jobs_upload_id;
-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
