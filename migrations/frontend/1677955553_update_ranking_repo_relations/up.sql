TRUNCATE codeintel_ranking_definitions;
TRUNCATE codeintel_ranking_path_counts_inputs;

ALTER TABLE codeintel_ranking_definitions DROP COLUMN IF EXISTS repository;
ALTER TABLE codeintel_ranking_path_counts_inputs DROP COLUMN IF EXISTS repository;
ALTER TABLE codeintel_ranking_path_counts_inputs ADD COLUMN IF NOT EXISTS repository_id INTEGER NOT NULL;

CREATE INDEX IF NOT EXISTS codeintel_ranking_path_counts_inputs_graph_key_and_repository_id ON codeintel_ranking_path_counts_inputs(graph_key, repository_id);
CREATE INDEX IF NOT EXISTS codeintel_ranking_path_counts_inputs_graph_key_repository_id_id ON codeintel_ranking_path_counts_inputs(graph_key, repository_id, id) INCLUDE (document_path) WHERE NOT processed;
