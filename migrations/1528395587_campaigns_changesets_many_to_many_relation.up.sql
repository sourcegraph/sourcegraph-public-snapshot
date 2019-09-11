BEGIN;

ALTER TABLE campaigns ADD COLUMN changeset_ids jsonb
NOT NULL DEFAULT '{}' CHECK (jsonb_typeof(changeset_ids) = 'object');

CREATE INDEX campaigns_changeset_ids_gin_idx ON campaigns
USING GIN (changeset_ids);

ALTER TABLE changesets DROP COLUMN campaign_id;
ALTER TABLE changesets ADD COLUMN campaign_ids jsonb
NOT NULL DEFAULT '{}' CHECK (jsonb_typeof(campaign_ids) = 'object');

CREATE INDEX changesets_campaign_ids_gin_idx ON changesets
USING GIN (campaign_ids);

COMMIT;
