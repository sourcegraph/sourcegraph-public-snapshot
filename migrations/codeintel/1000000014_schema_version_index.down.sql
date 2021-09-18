BEGIN;

DROP INDEX IF EXISTS lsif_data_documents_schema_versions_dump_id_schema_version_bounds;
DROP INDEX IF EXISTS lsif_data_definitions_schema_versions_dump_id_schema_version_bounds;
DROP INDEX IF EXISTS lsif_data_references_schema_versions_dump_id_schema_version_bounds;

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE codeintel_schema_migrations SET dirty = 'f'
COMMIT;
