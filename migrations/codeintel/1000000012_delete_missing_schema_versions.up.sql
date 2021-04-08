BEGIN;

-- Delete data from the schema version count tables that was inserted from the metadata table but does
-- not actually exist in the table that it shadows. This should only delete some records that we inserted
-- in 1000000007_definitions_locations_counts.up.sql and 1000000008_references_locations_counts.up.sql.

DELETE FROM lsif_data_documents_schema_versions sv WHERE NOT EXISTS (SELECT 1 FROM lsif_data_documents d WHERE d.dump_id = sv.dump_id);
DELETE FROM lsif_data_definitions_schema_versions sv WHERE NOT EXISTS (SELECT 1 FROM lsif_data_definitions d WHERE d.dump_id = sv.dump_id);
DELETE FROM lsif_data_references_schema_versions sv WHERE NOT EXISTS (SELECT 1 FROM lsif_data_references r WHERE r.dump_id = sv.dump_id);

COMMIT;
