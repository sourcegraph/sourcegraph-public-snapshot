ALTER TABLE gitserver_repos
    ADD COLUMN IF NOT EXISTS corrupted_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT '0001-01-01 00:00:00.0';

COMMENT ON COLUMN gitserver_repos.corrupted_at IS 'Timestamp of when repo corruption was detected';
