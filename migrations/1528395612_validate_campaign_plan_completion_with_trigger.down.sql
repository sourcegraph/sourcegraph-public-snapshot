BEGIN;

DROP TRIGGER IF EXISTS trig_validate_campaign_plan_is_finished ON campaigns;
DROP FUNCTION IF EXISTS validate_campaign_plan_is_finished();

DROP INDEX IF EXISTS campaign_jobs_campaign_plan_id;
DROP INDEX IF EXISTS campaign_jobs_started_at;
DROP INDEX IF EXISTS campaign_jobs_finished_at;

COMMIT;
