BEGIN;

ALTER TABLE campaign_jobs DROP CONSTRAINT campaign_jobs_repo_id_fkey,
  ADD CONSTRAINT campaign_jobs_repo_id_fkey
    FOREIGN KEY (repo_id)
    REFERENCES repo(id)
    DEFERRABLE INITIALLY IMMEDIATE;

COMMIT;
