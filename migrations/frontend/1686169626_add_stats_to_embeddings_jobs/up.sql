CREATE TABLE IF NOT EXISTS repo_embedding_job_stats (
    job_id INTEGER PRIMARY KEY REFERENCES repo_embedding_jobs(id) ON DELETE CASCADE DEFERRABLE,
    is_incremental BOOLEAN NOT NULL DEFAULT FALSE,
    code_files_total INTEGER NOT NULL DEFAULT 0,
    code_files_embedded INTEGER NOT NULL DEFAULT 0,
    code_chunks_embedded INTEGER NOT NULL DEFAULT 0,
    code_files_skipped JSONB NOT NULL DEFAULT '{}',
    code_bytes_embedded INTEGER NOT NULL DEFAULT 0,
    text_files_total INTEGER NOT NULL DEFAULT 0,
    text_files_embedded INTEGER NOT NULL DEFAULT 0,
    text_chunks_embedded INTEGER NOT NULL DEFAULT 0,
    text_files_skipped JSONB NOT NULL DEFAULT '{}',
    text_bytes_embedded INTEGER NOT NULL DEFAULT 0
);
