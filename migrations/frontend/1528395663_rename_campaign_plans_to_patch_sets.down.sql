BEGIN;

ALTER TABLE patch_sets RENAME TO campaign_plans;

ALTER TABLE campaign_jobs RENAME COLUMN patch_set_id TO campaign_plan_id;
ALTER TABLE campaigns RENAME COLUMN patch_set_id TO campaign_plan_id;

COMMIT;
