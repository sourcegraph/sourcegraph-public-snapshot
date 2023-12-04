CREATE TABLE IF NOT EXISTS codeintel_ranking_path_counts_inputs (
    id bigserial PRIMARY KEY NOT NULL,
    repository text NOT NULL,
    document_path text NOT NULL,
    count int NOT NULL,
    graph_key text NOT NULL,
    processed boolean NOT NULL DEFAULT false
);

CREATE INDEX IF NOT EXISTS codeintel_ranking_path_counts_inputs_graph_key_and_repository ON codeintel_ranking_path_counts_inputs (graph_key, repository);
