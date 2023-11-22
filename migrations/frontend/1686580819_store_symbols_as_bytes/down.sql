-- Drop new objects
DROP INDEX IF EXISTS codeintel_ranking_definitions_graph_key_symbol_checksum_search;
ALTER TABLE codeintel_ranking_references DROP COLUMN IF EXISTS symbol_checksums;
ALTER TABLE codeintel_ranking_definitions DROP COLUMN IF EXISTS symbol_checksum;

-- Recreate old index
CREATE INDEX IF NOT EXISTS codeintel_ranking_definitions_graph_key_symbol_search ON codeintel_ranking_definitions(graph_key, symbol_name, exported_upload_id, document_path);
