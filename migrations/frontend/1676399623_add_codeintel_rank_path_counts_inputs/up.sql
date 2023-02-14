CREATE TABLE IF NOT EXISTS codeintel_ranking_path_counts_inputs (
    repository text NOT NULL,
    document_root text NOT NULL,
    document_path text NOT NULL,
    count int NOT NULL,
    graph_key text NOT NULL,
    processed boolean NOT NULL DEFAULT false
);

