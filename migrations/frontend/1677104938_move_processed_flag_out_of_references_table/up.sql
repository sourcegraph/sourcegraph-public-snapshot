ALTER TABLE codeintel_ranking_references DROP COLUMN IF EXISTS processed;

CREATE TABLE IF NOT EXISTS codeintel_ranking_references_processed (
    id                              SERIAL PRIMARY KEY,
    graph_key                       TEXT NOT NULL,
    codeintel_ranking_reference_id  INT NOT NULL,

    CONSTRAINT fk_codeintel_ranking_reference FOREIGN KEY (codeintel_ranking_reference_id) REFERENCES codeintel_ranking_references(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS codeintel_ranking_references_processed_graph_key_codeintel_ranking_reference_id ON codeintel_ranking_references_processed(graph_key, codeintel_ranking_reference_id);
