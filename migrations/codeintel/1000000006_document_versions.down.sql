BEGIN;

DROP TABLE lsif_data_documents_schema_versions;
DROP TRIGGER lsif_data_documents_schema_versions_insert ON lsif_data_documents;
DROP FUNCTION update_lsif_data_documents_schema_versions_insert;

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE codeintel_schema_migrations SET dirty = 'f'
COMMIT;
