BEGIN;

DROP INDEX IF EXISTS gitserver_repos_cloned_status_idx;
DROP INDEX IF EXISTS gitserver_repos_not_cloned_status_idx;
DROP INDEX IF EXISTS gitserver_repos_cloning_status_idx;
CREATE INDEX IF NOT EXISTS gitserver_repos_clone_status_idx ON gitserver_repos (repo_id);

COMMIT;
