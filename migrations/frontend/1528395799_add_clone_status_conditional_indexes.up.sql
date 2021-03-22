BEGIN;

-- Having three partial indexes should be faster since when the condition is met the index used will be
-- smaller meaning it will be faster to scan.
CREATE INDEX IF NOT EXISTS gitserver_repos_cloned_status_idx ON gitserver_repos (repo_id) WHERE clone_status = 'cloned';
CREATE INDEX IF NOT EXISTS gitserver_repos_not_cloned_status_idx ON gitserver_repos (repo_id) WHERE clone_status = 'not_cloned';
CREATE INDEX IF NOT EXISTS gitserver_repos_cloning_status_idx ON gitserver_repos (repo_id) WHERE clone_status = 'cloning';

DROP INDEX IF EXISTS gitserver_repos_clone_status_idx;

COMMIT;
