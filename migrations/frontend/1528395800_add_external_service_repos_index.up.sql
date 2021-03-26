BEGIN;

CREATE INDEX external_service_repos_idx ON external_service_repos(external_service_id, repo_id);

COMMIT;
