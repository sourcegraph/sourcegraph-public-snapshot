CREATE INDEX IF NOT EXISTS gitserver_repo_size_bytes
    ON gitserver_repos USING btree (repo_size_bytes)
