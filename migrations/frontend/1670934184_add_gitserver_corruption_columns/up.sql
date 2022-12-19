ALTER TABLE gitserver_repos
    ADD COLUMN IF NOT EXISTS corrupted_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT '0001-01-01 00:00:00.0';
    ADD COLUMN IF NOT EXISTS corruption_log TEXT NOT NULL DEFAULT '';

COMMENT ON COLUMN gitserver_repos.corrupted_at IS 'Timestamp of when repo corruption was detected';
COMMENT ON COLUMN gitserver_repos.corruption_log IS 'log output of the corruption that was detected on the repo';
