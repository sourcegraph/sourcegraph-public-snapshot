ALTER TABLE gitserver_repos
    ADD COLUMN IF NOT EXISTS corrupted_at TIMESTAMP WITH TIME ZONE;

COMMENT ON COLUMN gitserver_repos.corrupted_at IS 'Timestamp of when repo corruption was dectected';
