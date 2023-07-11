ALTER TABLE IF EXISTS file_embedding_jobs
DROP CONSTRAINT IF EXISTS fk_embedding_plugin;

ALTER TABLE IF EXISTS file_embedding_jobs
RENAME COLUMN embedding_plugin_id TO archive_id;

ALTER TABLE IF EXISTS embedding_plugin_files
DROP CONSTRAINT IF EXISTS fk_embedding_plugin;

ALTER TABLE IF EXISTS embedding_plugin_files
ALTER COLUMN contents TYPE bytea USING contents::bytea;

DROP TABLE IF EXISTS embedding_plugins;
