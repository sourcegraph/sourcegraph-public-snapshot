BEGIN;

DROP INDEX IF EXISTS campaign_jobs_campaign_plan_id;
DROP INDEX IF EXISTS changeset_jobs_campaign_job_id;

COMMIT;
