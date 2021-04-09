BEGIN;

CREATE INDEX IF NOT EXISTS
    gitserver_repos_last_error_idx ON gitserver_repos(last_error) WHERE last_error IS NOT NULL;

COMMIT;
