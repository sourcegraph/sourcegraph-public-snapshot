DROP INDEX IF EXISTS codeintel_ranking_definitions_symbol_name;
DROP INDEX IF EXISTS codeintel_ranking_definitions_upload_id;
DROP INDEX IF EXISTS codeintel_ranking_path_counts_inputs_graph_key_and_repository_id;
DROP INDEX IF EXISTS codeintel_ranking_definitions_graph_key_id;
DROP INDEX IF EXISTS codeintel_ranking_references_graph_key_id;
DROP INDEX IF EXISTS codeintel_ranking_path_counts_inputs_graph_key_repository_id_id;
DROP INDEX IF EXISTS codeintel_path_ranks_updated_at;

TRUNCATE codeintel_ranking_definitions CASCADE;
TRUNCATE codeintel_ranking_references CASCADE;
TRUNCATE codeintel_ranking_path_counts_inputs CASCADE;
TRUNCATE codeintel_path_ranks CASCADE;

CREATE INDEX IF NOT EXISTS codeintel_ranking_definitions_graph_key_symbol_search ON codeintel_ranking_definitions(graph_key, symbol_name, upload_id, document_path);
CREATE INDEX IF NOT EXISTS codeintel_ranking_references_graph_key_id ON codeintel_ranking_references(graph_key, id);
CREATE INDEX IF NOT EXISTS codeintel_ranking_path_counts_inputs_graph_key_repository_id_id ON codeintel_ranking_path_counts_inputs(graph_key, repository_id, id) WHERE NOT processed;
CREATE INDEX IF NOT EXISTS codeintel_path_ranks_graph_key ON codeintel_path_ranks(graph_key, updated_at NULLS FIRST, id);
CREATE INDEX IF NOT EXISTS codeintel_path_ranks_repository_id_updated_at_id ON codeintel_path_ranks(repository_id, updated_at NULLS FIRST, id);
