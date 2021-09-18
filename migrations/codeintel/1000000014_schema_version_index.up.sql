BEGIN;

CREATE INDEX IF NOT EXISTS lsif_data_documents_schema_versions_dump_id_schema_version_bounds ON lsif_data_documents_schema_versions (dump_id, min_schema_version, max_schema_version);
CREATE INDEX IF NOT EXISTS lsif_data_definitions_schema_versions_dump_id_schema_version_bounds ON lsif_data_definitions_schema_versions (dump_id, min_schema_version, max_schema_version);
CREATE INDEX IF NOT EXISTS lsif_data_references_schema_versions_dump_id_schema_version_bounds ON lsif_data_references_schema_versions (dump_id, min_schema_version, max_schema_version);

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE codeintel_schema_migrations SET dirty = 'f'
COMMIT;
