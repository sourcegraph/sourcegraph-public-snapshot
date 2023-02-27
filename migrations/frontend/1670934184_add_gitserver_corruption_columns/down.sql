ALTER TABLE gitserver_repos
    DROP COLUMN IF EXISTS corrupted_at,
    DROP COLUMN IF EXISTS corrupted_logs;

