UPDATE codeintel_ranking_path_counts_inputs SET processed = true WHERE NOT processed;

CREATE UNIQUE INDEX IF NOT EXISTS codeintel_ranking_path_counts_inputs_graph_key_unique_definition_id ON codeintel_ranking_path_counts_inputs(graph_key, definition_id) WHERE NOT processed;
