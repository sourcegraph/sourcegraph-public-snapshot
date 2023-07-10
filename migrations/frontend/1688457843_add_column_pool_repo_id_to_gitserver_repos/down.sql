ALTER TABLE gitserver_repos
	DROP COLUMN IF EXISTS pool_repo_id;

DROP INDEX IF EXISTS gitserver_repos_pool_repo_id;
