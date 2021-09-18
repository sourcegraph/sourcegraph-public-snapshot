BEGIN;

ALTER TABLE lsif_data_documentation_mappings DROP COLUMN file_path;

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE codeintel_schema_migrations SET dirty = 'f'
COMMIT;
