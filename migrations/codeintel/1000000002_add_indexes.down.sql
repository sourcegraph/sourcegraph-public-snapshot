BEGIN;

DROP INDEX IF EXISTS lsif_data_metadata_temp;
DROP INDEX IF EXISTS lsif_data_documents_temp;
DROP INDEX IF EXISTS lsif_data_result_chunks_temp;
DROP INDEX IF EXISTS lsif_data_definitions_temp;
DROP INDEX IF EXISTS lsif_data_references_temp;

COMMIT;
