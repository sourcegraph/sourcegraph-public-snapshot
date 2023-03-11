CREATE TABLE IF NOT EXISTS codeintel_ranking_definitions (
    id bigserial PRIMARY KEY NOT NULL,
    upload_id int NOT NULL,
    symbol_name text NOT NULL,
    repository text NOT NULL,
    document_path text NOT NULL,
    graph_key text NOT NULL
);

CREATE INDEX IF NOT EXISTS codeintel_ranking_definitions_upload_id ON codeintel_ranking_definitions (upload_id);
CREATE INDEX IF NOT EXISTS codeintel_ranking_definitions_symbol_name ON codeintel_ranking_definitions (symbol_name);

CREATE TABLE IF NOT EXISTS codeintel_ranking_references (
    id bigserial PRIMARY KEY NOT NULL,
    upload_id int NOT NULL,
    symbol_names text[] NOT NULL,
    graph_key text NOT NULL,
    processed boolean NOT NULL DEFAULT false
);

COMMENT ON TABLE codeintel_ranking_references IS 'References for a given upload proceduced by background job consuming SCIP indexes.';

CREATE INDEX IF NOT EXISTS codeintel_ranking_references_upload_id ON codeintel_ranking_references (upload_id);
