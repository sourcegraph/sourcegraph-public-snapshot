BEGIN;

DROP TRIGGER IF EXISTS trig_validate_campaign_plan_is_finished ON campaigns;
DROP FUNCTION IF EXISTS validate_campaign_plan_is_finished();

ALTER TABLE campaign_jobs DROP COLUMN IF EXISTS error;
ALTER TABLE campaign_jobs DROP COLUMN IF EXISTS started_at;
ALTER TABLE campaign_jobs DROP COLUMN IF EXISTS finished_at;
ALTER TABLE campaign_jobs DROP COLUMN IF EXISTS description;


COMMIT;
