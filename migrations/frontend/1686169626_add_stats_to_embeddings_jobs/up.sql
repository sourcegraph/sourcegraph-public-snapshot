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

ALTER TABLE repo_embedding_jobs
ADD COLUMN stat_has_ranks BOOLEAN NOT NULL DEFAULT FALSE,
ADD COLUMN stat_is_incremental BOOLEAN NOT NULL DEFAULT FALSE,
ADD COLUMN stat_code_files_total INTEGER NOT NULL DEFAULT 0,
ADD COLUMN stat_code_files_embedded INTEGER NOT NULL DEFAULT 0,
ADD COLUMN stat_code_chunks_embedded INTEGER NOT NULL DEFAULT 0,
ADD COLUMN stat_code_files_skipped JSONB NOT NULL DEFAULT '{}',
ADD COLUMN stat_code_bytes_skipped JSONB NOT NULL DEFAULT '{}',
ADD COLUMN stat_code_bytes_embedded INTEGER NOT NULL DEFAULT 0,
ADD COLUMN stat_text_files_total INTEGER NOT NULL DEFAULT 0,
ADD COLUMN stat_text_files_embedded INTEGER NOT NULL DEFAULT 0,
ADD COLUMN stat_text_chunks_embedded INTEGER NOT NULL DEFAULT 0,
ADD COLUMN stat_text_files_skipped JSONB NOT NULL DEFAULT '{}',
ADD COLUMN stat_text_bytes_skipped JSONB NOT NULL DEFAULT '{}',
ADD COLUMN stat_text_bytes_embedded INTEGER NOT NULL DEFAULT 0;
