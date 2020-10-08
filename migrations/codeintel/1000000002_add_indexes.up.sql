BEGIN;

CREATE INDEX IF NOT EXISTS lsif_data_metadata_temp ON lsif_data_metadata (dump_id);
CREATE INDEX IF NOT EXISTS lsif_data_documents_temp ON lsif_data_documents (dump_id, path);
CREATE INDEX IF NOT EXISTS lsif_data_result_chunks_temp ON lsif_data_result_chunks (dump_id, idx);
CREATE INDEX IF NOT EXISTS lsif_data_definitions_temp ON lsif_data_definitions (dump_id, scheme, identifier);
CREATE INDEX IF NOT EXISTS lsif_data_references_temp ON lsif_data_references (dump_id, scheme, identifier);

COMMIT;
