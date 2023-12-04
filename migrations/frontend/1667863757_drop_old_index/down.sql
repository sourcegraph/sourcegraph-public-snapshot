CREATE INDEX IF NOT EXISTS codeintel_path_rank_graph_key_id_repository_name_processed ON codeintel_path_rank_inputs(graph_key, id, repository_name)
WHERE
    NOT processed;
