DROP FUNCTION IF EXISTS addr_for_repo;
DROP VIEW IF EXISTS repo_update_jobs_with_repo_name;
DROP TABLE IF EXISTS repo_update_jobs;

ALTER TABLE IF EXISTS gitserver_repos
    ADD COLUMN IF NOT EXISTS cloning_progress text DEFAULT '';
