ALTER TABLE codeintel_ranking_path_counts_inputs DROP COLUMN IF EXISTS document_path;
ALTER TABLE codeintel_ranking_path_counts_inputs DROP COLUMN IF EXISTS repository_id;
ALTER TABLE codeintel_ranking_path_counts_inputs ADD COLUMN IF NOT EXISTS definition_id BIGINT;

CREATE INDEX IF NOT EXISTS codeintel_ranking_path_counts_inputs_graph_key_definition_id ON codeintel_ranking_path_counts_inputs(graph_key, definition_id, id) WHERE NOT processed;
