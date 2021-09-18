BEGIN;

DROP INDEX IF EXISTS lsif_indexes_repository_id_commit;

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
