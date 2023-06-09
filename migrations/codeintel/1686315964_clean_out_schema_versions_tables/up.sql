DROP TABLE IF EXISTS codeintel_scip_documents_schema_versions;

-- Clear data that we've been neglecting to clean up
DELETE FROM codeintel_scip_symbols_schema_versions sv         WHERE NOT EXISTS (SELECT 1 FROM codeintel_scip_metadata m WHERE m.upload_id = sv.upload_id);
DELETE FROM codeintel_scip_document_lookup_schema_versions sv WHERE NOT EXISTS (SELECT 1 FROM codeintel_scip_metadata m WHERE m.upload_id = sv.upload_id);
