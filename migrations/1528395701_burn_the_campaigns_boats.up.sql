BEGIN;

-- We're pruning pre-spec campaigns and changesets so that we can tighten up
-- the schema to match the new spec-based way campaigns are represented.
-- However, we need to keep the data so that we can provide a somewhat graceful
-- migration, so the first thing we'll do is to create tables to retain old
-- campaigns and changesets.

CREATE TABLE IF NOT EXISTS
    campaigns_old
AS
    SELECT
        *
    FROM
        campaigns
    WHERE
        campaign_spec_id IS NULL;

-- This query's a bit tricky: what we want is changesets that are _only_
-- attached to campaigns without specs. Changesets that are (somehow) attached
-- to both a spec and non-spec campaign should be left untouched.

CREATE TABLE IF NOT EXISTS
    changesets_old
AS
    SELECT
        *
    FROM
        changesets
    WHERE
        campaign_ids ?| array(
            SELECT
                id::VARCHAR
            FROM
                campaigns
            WHERE campaign_spec_id IS NULL
        )
        AND NOT campaign_ids ?| array(
            SELECT
                id::VARCHAR
            FROM
                campaigns
            WHERE campaign_spec_id IS NOT NULL
        );

-- Now we've set up the tables, we can take the next step, which is to delete
-- the old campaigns and changesets.

DELETE FROM
    changesets
WHERE
    campaign_ids ?| array(
        SELECT
            id::VARCHAR
        FROM
            campaigns
        WHERE campaign_spec_id IS NULL
    )
    AND NOT campaign_ids ?| array(
        SELECT
            id::VARCHAR
        FROM
            campaigns
        WHERE campaign_spec_id IS NOT NULL
    );

DELETE FROM
    campaigns
WHERE
    campaign_spec_id IS NULL;

-- Now we can alter our fields.

ALTER TABLE campaigns
    ALTER COLUMN last_applier_id SET NOT NULL,
    ALTER COLUMN last_applied_at SET NOT NULL,
    ALTER COLUMN campaign_spec_id SET NOT NULL;

COMMIT;
