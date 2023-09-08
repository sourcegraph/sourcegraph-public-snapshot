-- Perform migration here.
--
-- See /migrations/README.md. Highlights:
--  * Make migrations idempotent (use IF EXISTS)
--  * Make migrations backwards-compatible (old readers/writers must continue to work)
--  * If you are using CREATE INDEX CONCURRENTLY, then make sure that only one statement
--    is defined per file, and that each such statement is NOT wrapped in a transaction.
--    Each such migration must also declare "createIndexConcurrently: true" in their
--    associated metadata.yaml file.
--  * If you are modifying Postgres extensions, you must also declare "privileged: true"
--    in the associated metadata.yaml file.
CREATE TABLE IF NOT EXISTS file_metrics (
    id SERIAL PRIMARY KEY,
    repo_id integer NOT NULL REFERENCES repo(id) ON DELETE CASCADE DEFERRABLE,
    file_path integer NOT NULL,
    commit_sha bytea NOT NULL DEFAULT 'HEAD',
    size_in_bytes int NOT NULL DEFAULT 0,
    line_count int NOT NULL DEFAULT 0,
    word_count int NOT NULL DEFAULT 0,
    languages text[]
);

CREATE UNIQUE INDEX IF NOT EXISTS file_metrics_id_unique ON file_metrics USING btree (repo_id, file_path, commit_sha);
