CREATE INDEX CONCURRENTLY IF NOT EXISTS repo_name_case_sensitive_trgm_idx ON repo USING gin ((name::text) gin_trgm_ops);
