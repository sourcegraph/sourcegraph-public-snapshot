ALTER TABLE repo_embedding_job_stats
    ADD COLUMN IF NOT EXISTS code_chunks_excluded INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS text_chunks_excluded INTEGER NOT NULL DEFAULT 0;
