BEGIN;

-- Add and populate branch column to the campaigns table.
-- The branch name is generated with this pattern:
--
-- sourcegraph/campaign-<unix creation date>
--
-- This pattern is based from the default branch name 
-- in previous versions.
ALTER TABLE campaigns ADD COLUMN branch text;
UPDATE 
    campaigns 
SET 
    branch=concat('sourcegraph/campaign-', date_part('epoch', created_at)::int);
ALTER TABLE campaigns ALTER branch SET NOT NULL;

-- Add and populate branch column to the changeset jobs tables.
-- The branch name is inherited from the campaign if the changeset
-- job is finished running.
ALTER TABLE changeset_jobs ADD COLUMN branch text;
UPDATE changeset_jobs AS csj 
    SET branch=c.branch
FROM campaigns c
WHERE csj.campaign_id = c.id AND csj.finished_at IS NOT NULL;

COMMIT;
