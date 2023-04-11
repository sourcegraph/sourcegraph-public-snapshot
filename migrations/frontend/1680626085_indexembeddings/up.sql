CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS embedding_versions (
    id SERIAL PRIMARY KEY,
    repo_id INTEGER NOT NULL REFERENCES repo(id),
    revision TEXT NOT NULL,
    UNIQUE(repo_id, revision)
);

CREATE TABLE IF NOT EXISTS text_embeddings (
    id SERIAL PRIMARY KEY,
    version_id INTEGER NOT NULL REFERENCES embedding_versions(id),
    embedding vector(768) NOT NULL,
    file_name TEXT NOT NULL,
    start_line INTEGER NOT NULL,
    end_line INTEGER NOT NULL,
    rank FLOAT NOT NULL
);

CREATE TABLE IF NOT EXISTS code_embeddings (
    id SERIAL PRIMARY KEY,
    version_id INTEGER NOT NULL REFERENCES embedding_versions(id),
    embedding vector(768) NOT NULL,
    file_name TEXT NOT NULL,
    start_line INTEGER NOT NULL,
    end_line INTEGER NOT NULL,
    rank FLOAT NOT NULL
);
