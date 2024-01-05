CREATE INDEX IF NOT EXISTS idx_repo_gitlab_topics
ON repo
USING GIN((metadata->'topics'))
WHERE external_service_type = 'gitlab';

CREATE INDEX IF NOT EXISTS idx_repo_github_topics
ON repo
USING GIN((metadata->'RepositoryTopics'->'Nodes'))
WHERE external_service_type = 'github';
