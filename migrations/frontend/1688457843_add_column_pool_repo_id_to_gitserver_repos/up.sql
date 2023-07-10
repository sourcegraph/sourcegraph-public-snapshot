ALTER TABLE gitserver_repos
	ADD COLUMN IF NOT EXISTS pool_repo_id INTEGER DEFAULT NULL;

COMMENT ON COLUMN gitserver_repos.pool_repo_id IS 'This is used to refer to the pool repository for deduplicated repos';

CREATE INDEX IF NOT EXISTS gitserver_repos_pool_repo_id ON gitserver_repos (pool_repo_id);
