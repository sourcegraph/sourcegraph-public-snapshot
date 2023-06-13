-- Drop reliance on symbol names
DROP INDEX IF EXISTS codeintel_ranking_definitions_graph_key_symbol_search;

-- Add symbol checksums columns
ALTER TABLE codeintel_ranking_references ADD COLUMN IF NOT EXISTS symbol_checksums bytea[] NOT NULL DEFAULT '{}';
ALTER TABLE codeintel_ranking_definitions ADD COLUMN IF NOT EXISTS symbol_checksum bytea NOT NULL DEFAULT ''::bytea;
CREATE INDEX IF NOT EXISTS codeintel_ranking_definitions_graph_key_symbol_checksum_search ON codeintel_ranking_definitions(graph_key, symbol_checksum, exported_upload_id, document_path);
