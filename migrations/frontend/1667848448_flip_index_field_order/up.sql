-- Effectively replaces codeintel_path_rank_graph_key_id_repository_name_processed
CREATE INDEX IF NOT EXISTS codeintel_path_rank_inputs_graph_key_repository_name_id_processed ON codeintel_path_rank_inputs(graph_key, repository_name, id)
WHERE
    NOT processed;
