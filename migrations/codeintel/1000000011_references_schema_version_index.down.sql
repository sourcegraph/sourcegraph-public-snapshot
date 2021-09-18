BEGIN;

DROP INDEX IF EXISTS lsif_data_references_dump_id_schema_version;

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE codeintel_schema_migrations SET dirty = 'f'
COMMIT;
