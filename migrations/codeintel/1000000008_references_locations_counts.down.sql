BEGIN;

DROP TABLE lsif_data_references_schema_versions;
DROP TRIGGER lsif_data_references_schema_versions_insert ON lsif_data_references;
DROP FUNCTION update_lsif_data_references_schema_versions_insert;

ALTER TABLE lsif_data_references DROP COLUMN IF EXISTS schema_version;
ALTER TABLE lsif_data_references DROP COLUMN IF EXISTS num_locations;

COMMIT;
