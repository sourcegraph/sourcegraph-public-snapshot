CREATE INDEX IF NOT EXISTS codeintel_scip_symbol_names_upload_id_roots ON codeintel_scip_symbol_names(upload_id) WHERE prefix_id IS NULL;
CREATE INDEX IF NOT EXISTS codeisdntel_scip_symbol_names_upload_id_children ON codeintel_scip_symbol_names(upload_id, prefix_id) WHERE prefix_id IS NOT NULL;
