ALTER TABLE gitserver_repos
    ADD COLUMN IF NOT EXISTS corrupted_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN IF NOT EXISTS corruption_logs JSONB NOT NULL DEFAULT '[]';

COMMENT ON COLUMN gitserver_repos.corrupted_at IS 'Timestamp of when repo corruption was detected';
COMMENT ON COLUMN gitserver_repos.corruption_logs IS 'Log output of repo corruptions that have been detected - encoded as json';
