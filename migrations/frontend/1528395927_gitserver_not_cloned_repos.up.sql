BEGIN;

CREATE INDEX IF NOT EXISTS gitserver_repos_not_uncloned_idx ON gitserver_repos (repo_id) WHERE clone_status <> 'not_cloned'::text;

COMMIT;
