ALTER TABLE repo_embedding_jobs
    ADD COLUMN IF NOT EXISTS model_id TEXT NOT NULL DEFAULT '';
