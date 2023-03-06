TRUNCATE codeintel_ranking_definitions;
TRUNCATE codeintel_ranking_path_counts_inputs;

ALTER TABLE codeintel_ranking_definitions ADD COLUMN IF NOT EXISTS repository TEXT NOT NULL;
ALTER TABLE codeintel_ranking_path_counts_inputs DROP COLUMN IF EXISTS repository_id;
ALTER TABLE codeintel_ranking_path_counts_inputs ADD COLUMN IF NOT EXISTS repository TEXT NOT NULL;

CREATE INDEX IF NOT EXISTS codeintel_ranking_path_counts_inputs_graph_key_and_repository ON codeintel_ranking_path_counts_inputs(graph_key, repository);
CREATE INDEX IF NOT EXISTS codeintel_ranking_path_counts_inputs_graph_key_repository_id_pr ON codeintel_ranking_path_counts_inputs(graph_key, repository, id) INCLUDE (document_path) WHERE NOT processed;
