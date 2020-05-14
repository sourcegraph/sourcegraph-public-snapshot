BEGIN;

ALTER TABLE changesets ADD COLUMN created_by_campaign BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE changesets ADD COLUMN added_to_campaign BOOLEAN NOT NULL DEFAULT false;

UPDATE changesets as cs
SET added_to_campaign = true
WHERE NOT EXISTS
    (SELECT 1 FROM changeset_jobs WHERE changeset_id = cs.id AND cs.campaign_ids::text <> '' AND cs.campaign_ids::text <> '{}');

UPDATE changesets as cs
SET created_by_campaign = true
WHERE EXISTS
    (SELECT 1 FROM changeset_jobs WHERE changeset_id = cs.id);

COMMIT;
