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

CREATE TABLE IF NOT EXISTS repo_embedding_job_stats (
    job_id INTEGER PRIMARY KEY REFERENCES (repo_embedding_jobs.id) ON DELETE CASCADE DEFERRABLE,
    stat_has_ranks BOOLEAN NOT NULL DEFAULT FALSE,
    stat_is_incremental BOOLEAN NOT NULL DEFAULT FALSE,
    stat_code_files_total INTEGER NOT NULL DEFAULT 0,
    stat_code_files_embedded INTEGER NOT NULL DEFAULT 0,
    stat_code_chunks_embedded INTEGER NOT NULL DEFAULT 0,
    stat_code_files_skipped JSONB NOT NULL DEFAULT '{}',
    stat_code_bytes_skipped JSONB NOT NULL DEFAULT '{}',
    stat_code_bytes_embedded INTEGER NOT NULL DEFAULT 0,
    stat_text_files_total INTEGER NOT NULL DEFAULT 0,
    stat_text_files_embedded INTEGER NOT NULL DEFAULT 0,
    stat_text_chunks_embedded INTEGER NOT NULL DEFAULT 0,
    stat_text_files_skipped JSONB NOT NULL DEFAULT '{}',
    stat_text_bytes_skipped JSONB NOT NULL DEFAULT '{}',
    stat_text_bytes_embedded INTEGER NOT NULL DEFAULT 0
);
