ALTER TABLE codeintel_ranking_path_counts_inputs DROP COLUMN IF EXISTS definition_id;
ALTER TABLE codeintel_ranking_path_counts_inputs ADD COLUMN IF NOT EXISTS document_path TEXT NOT NULL;
ALTER TABLE codeintel_ranking_path_counts_inputs ADD COLUMN IF NOT EXISTS repository_id INTEGER NOT NULL;

CREATE INDEX IF NOT EXISTS codeintel_ranking_path_counts_inputs_graph_key_repository_id_id ON codeintel_ranking_path_counts_inputs(graph_key, repository_id, id) WHERE NOT processed;
