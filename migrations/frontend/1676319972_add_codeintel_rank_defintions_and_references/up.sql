CREATE TABLE codeintel_ranking_definitions (
    upload_id int NOT NULL,
    symbol_name text NOT NULL,
    repository text NOT NULL,
    document_root text NOT NULL,
    document_path text NOT NULL
);

CREATE INDEX IF NOT EXISTS codeintel_ranking_definitions_upload_id ON codeintel_ranking_definitions (upload_id);
CREATE INDEX IF NOT EXISTS codeintel_ranking_definitions_symbol_name ON codeintel_ranking_definitions (symbol_name);

CREATE TABLE codeintel_ranking_references (
    upload_id int NOT NULL,
    symbol_names text[] NOT NULL
);
