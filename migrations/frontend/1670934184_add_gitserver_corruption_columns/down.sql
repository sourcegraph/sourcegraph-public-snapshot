ALTER TABLE gitserver_repos
    DROP COLUMN IF EXISTS corrupted_at;

ALTER TABLE gitserver_repos
    DROP COLUMN IF EXISTS corrupted_logs;

ALTER TABLE gitserver_repos DROP COLUMN IF EXISTS corruption_logs;
