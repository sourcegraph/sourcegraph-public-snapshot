CREATE INDEX CONCURRENTLY IF NOT EXISTS codeintel_ranking_path_counts_inputs_graph_key_repository_id_processed
ON codeintel_ranking_path_counts_inputs (graph_key, repository, id)
INCLUDE (document_path)
WHERE NOT processed;
