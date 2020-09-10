BEGIN;

ALTER TABLE campaign_plans RENAME TO patch_sets;

ALTER TABLE campaign_jobs RENAME COLUMN campaign_plan_id TO patch_set_id;
ALTER TABLE campaigns RENAME COLUMN campaign_plan_id TO patch_set_id;

COMMIT;
