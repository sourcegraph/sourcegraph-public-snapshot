CREATE TABLE IF NOT EXISTS codeintel_initial_path_ranks(
    id BIGSERIAL PRIMARY KEY NOT NULL,
    repository_id INTEGER NOT NULL,
    document_path TEXT NOT NULL,
    graph_key   TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS codeintel_initial_path_ranks_processed(
    id BIGSERIAL PRIMARY KEY,
    graph_key TEXT NOT NULL,
    codeintel_initial_path_ranks_id BIGINT NOT NULL,

    CONSTRAINT fk_codeintel_initial_path_ranks FOREIGN KEY (codeintel_initial_path_ranks_id) REFERENCES codeintel_initial_path_ranks(id) ON DELETE CASCADE
);


CREATE INDEX IF NOT EXISTS codeintel_initial_path_ranks_graph_key_and_repository_id ON codeintel_initial_path_ranks (graph_key, repository_id);
