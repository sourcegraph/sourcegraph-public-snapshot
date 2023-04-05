CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS repo_embeddings (
    id SERIAL PRIMARY KEY,
    repo_id INTEGER NOT NULL REFERENCES repo(id),
    revision TEXT NOT NULL,
    code_index vector(1536) NOT NULL,
    code_column_dimension INTEGER NOT NULL,
    code_row_metadata JSONB NOT NULL,
    code_ranks FLOAT[],
    text_index vector(1536) NOT NULL,
    text_column_dimension INTEGER NOT NULL,
    text_row_metadata JSONB NOT NULL,
    text_ranks FLOAT[]
); --  PARTITION BY HASH (repo_id)?
