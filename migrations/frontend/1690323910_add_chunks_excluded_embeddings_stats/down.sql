ALTER TABLE repo_embedding_job_stats
    DROP COLUMN IF EXISTS code_chunks_excluded,
    DROP COLUMN IF EXISTS text_chunks_excluded;
