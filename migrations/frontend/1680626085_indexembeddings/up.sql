CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS embedding_vectors (
    id SERIAL PRIMARY KEY,
    repo_id INTEGER NOT NULL REFERENCES repo(id),
    embedding vector(1536) NOT NULL
);

CREATE TABLE IF NOT EXISTS text_embeddings (
    embedding_id INTEGER NOT NULL REFERENCES embedding_vectors(id),
    repo_id INTEGER NOT NULL REFERENCES repo(id),
    revision TEXT NOT NULL,
    file_name TEXT NOT NULL,
    start_line INTEGER NOT NULL,
    end_line INTEGER NOT NULL,
    rank FLOAT NOT NULL
);

CREATE TABLE IF NOT EXISTS code_embeddings (
    embedding_id INTEGER NOT NULL REFERENCES embedding_vectors(id),
    repo_id INTEGER NOT NULL REFERENCES repo(id),
    revision TEXT NOT NULL,
    file_name TEXT NOT NULL,
    start_line INTEGER NOT NULL,
    end_line INTEGER NOT NULL,
    rank FLOAT NOT NULL
);
