-- Undo the changes made in the up migration

COMMENT ON INDEX idx_repo_github_topics
IS NULL;
