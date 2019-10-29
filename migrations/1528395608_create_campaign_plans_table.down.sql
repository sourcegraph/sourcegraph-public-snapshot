BEGIN;

ALTER TABLE campaigns DROP COLUMN IF EXISTS campaign_plan_id;

DROP TABLE IF EXISTS campaign_plans;

COMMIT;
