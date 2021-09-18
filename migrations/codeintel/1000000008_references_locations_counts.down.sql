BEGIN;

DROP TABLE lsif_data_references_schema_versions;
DROP TRIGGER lsif_data_references_schema_versions_insert ON lsif_data_references;
DROP FUNCTION update_lsif_data_references_schema_versions_insert;

ALTER TABLE lsif_data_references DROP COLUMN IF EXISTS schema_version;
ALTER TABLE lsif_data_references DROP COLUMN IF EXISTS num_locations;

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE codeintel_schema_migrations SET dirty = 'f'
COMMIT;
