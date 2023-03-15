CREATE INDEX CONCURRENTLY IF NOT EXISTS gitserver_repos_last_changed_idx ON gitserver_repos(last_changed, repo_id);
