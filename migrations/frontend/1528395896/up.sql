CREATE INDEX CONCURRENTLY gitserver_repos_last_error_idx ON gitserver_repos(repo_id) WHERE last_error IS NOT NULL;
