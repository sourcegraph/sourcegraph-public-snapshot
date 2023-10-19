DROP INDEX IF EXISTS idx_repo_github_topics;
DROP INDEX IF EXISTS idx_repo_gitlab_topics;
COMMENT ON idx_repo_topics IS 'For efficiently looking up repositories by topic';
