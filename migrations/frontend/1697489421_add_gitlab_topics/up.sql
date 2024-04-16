CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_repo_gitlab_topics
ON repo
USING GIN((metadata->'topics'))
WHERE external_service_type = 'gitlab';
