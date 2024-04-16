CREATE INDEX IF NOT EXISTS repo_embedding_jobs_repo ON repo_embedding_jobs USING btree (repo_id, revision)
