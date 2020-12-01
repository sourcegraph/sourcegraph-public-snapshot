BEGIN;

CREATE INDEX IF NOT EXISTS external_service_repos_repo_id ON external_service_repos(repo_id);

COMMIT;
