CREATE INDEX IF NOT EXISTS exhaustive_search_jobs_state ON exhaustive_search_jobs (state);
CREATE INDEX IF NOT EXISTS exhaustive_search_repo_jobs_state ON exhaustive_search_repo_jobs (state);
CREATE INDEX IF NOT EXISTS exhaustive_search_repo_revision_jobs_state ON exhaustive_search_repo_revision_jobs (state);
