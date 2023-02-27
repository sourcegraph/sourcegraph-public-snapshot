CREATE INDEX CONCURRENTLY IF NOT EXISTS gitserver_repos_not_explicitly_cloned_idx ON gitserver_repos (repo_id) WHERE clone_status <> 'cloned';
