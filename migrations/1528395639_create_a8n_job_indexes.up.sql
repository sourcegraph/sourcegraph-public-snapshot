BEGIN;

CREATE INDEX campaign_jobs_campaign_plan_id ON campaign_jobs (campaign_plan_id);
CREATE INDEX changeset_jobs_campaign_job_id ON changeset_jobs (campaign_job_id);

COMMIT;
