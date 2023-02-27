CREATE INDEX CONCURRENTLY IF NOT EXISTS repo_description_trgm_idx ON repo USING GIN (lower(description) gin_trgm_ops);
