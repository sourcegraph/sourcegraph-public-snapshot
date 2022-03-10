ALTER TABLE IF EXISTS gitserver_repos
    ADD COLUMN IF NOT EXISTS repo_size_bytes BIGINT;
