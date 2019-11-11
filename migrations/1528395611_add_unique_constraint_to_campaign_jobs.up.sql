BEGIN;

ALTER TABLE campaign_jobs ADD CONSTRAINT campaign_jobs_campaign_plan_repo_rev_unique
  UNIQUE (campaign_plan_id, repo_id, rev)
  DEFERRABLE INITIALLY IMMEDIATE;

COMMIT;
