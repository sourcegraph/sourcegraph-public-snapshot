-- +++
-- parent: 1528395893
-- +++

-- This speeds up IterateRepoGitserverStatus
CREATE INDEX CONCURRENTLY IF NOT EXISTS gitserver_repos_shard_id ON gitserver_repos(shard_id, repo_id);
