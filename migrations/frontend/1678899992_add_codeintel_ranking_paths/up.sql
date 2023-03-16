CREATE TABLE IF NOT EXISTS codeintel_initial_path_ranks(
    id               BIGSERIAL PRIMARY KEY NOT NULL,
    upload_id        INTEGER NOT NULL,
    document_path    TEXT NOT NULL,
    graph_key        TEXT NOT NULL,
    last_scanned_at  TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS codeintel_initial_path_ranks_graph_key_id ON codeintel_initial_path_ranks(graph_key, id);
CREATE INDEX IF NOT EXISTS codeintel_initial_path_ranks_graph_key_last_scanned_at ON codeintel_initial_path_ranks(graph_key, last_scanned_at NULLS FIRST, id);
CREATE INDEX IF NOT EXISTS codeintel_initial_path_upload_id ON codeintel_initial_path_ranks(upload_id);

CREATE TABLE IF NOT EXISTS codeintel_initial_path_ranks_processed(
    id                               BIGSERIAL PRIMARY KEY,
    graph_key                        TEXT NOT NULL,
    codeintel_initial_path_ranks_id  BIGINT NOT NULL,

    CONSTRAINT fk_codeintel_initial_path_ranks FOREIGN KEY (codeintel_initial_path_ranks_id) REFERENCES codeintel_initial_path_ranks(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS codeintel_initial_path_ranks_processed_cgraph_key_codeintel_initial_path_ranks_id ON codeintel_initial_path_ranks_processed(graph_key, codeintel_initial_path_ranks_id);
