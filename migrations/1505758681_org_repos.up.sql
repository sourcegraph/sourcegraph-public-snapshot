ALTER TABLE local_repos RENAME TO org_repos;
ALTER TABLE threads RENAME COLUMN local_repo_id TO org_repo_id;