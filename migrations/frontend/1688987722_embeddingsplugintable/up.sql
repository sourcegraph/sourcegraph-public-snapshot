CREATE TABLE IF NOT EXISTS embedding_plugins (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    original_source_url TEXT
);

ALTER TABLE IF EXISTS embedding_plugin_files
ADD CONSTRAINT fk_embedding_plugin
    FOREIGN KEY(embedding_plugin_id)
        REFERENCES embedding_plugins(id);

ALTER TABLE IF EXISTS embedding_plugin_files
ALTER COLUMN contents TYPE text;

ALTER TABLE IF EXISTS file_embedding_jobs
RENAME COLUMN archive_id TO embedding_plugin_id;

ALTER TABLE IF EXISTS file_embedding_jobs
ALTER COLUMN embedding_plugin_id TYPE integer USING embedding_plugin_id::integer;

ALTER TABLE IF EXISTS file_embedding_jobs
ADD CONSTRAINT fk_embedding_plugin
    FOREIGN KEY(embedding_plugin_id)
        REFERENCES embedding_plugins(id);
