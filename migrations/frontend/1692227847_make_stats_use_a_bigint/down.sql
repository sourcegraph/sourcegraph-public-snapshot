ALTER TABLE repo_embedding_job_stats
    ALTER COLUMN code_bytes_embedded TYPE integer,
    ALTER COLUMN text_bytes_embedded TYPE integer;
