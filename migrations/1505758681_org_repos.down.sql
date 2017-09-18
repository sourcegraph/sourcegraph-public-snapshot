ALTER TABLE org_repos RENAME TO local_repos;
ALTER TABLE threads RENAME COLUMN org_repo_id TO local_repo_id;