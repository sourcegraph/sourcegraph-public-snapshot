BEGIN;

ALTER TABLE campaign_jobs DROP CONSTRAINT campaign_jobs_campaign_plan_repo_rev_unique;

COMMIT;
