BEGIN;

CREATE INDEX IF NOT EXISTS
    gitserver_repos_last_error_idx ON gitserver_repos(last_error) WHERE last_error IS NOT NULL;

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
