BEGIN;

-- When we delete a `changeset` we remove its ID from the `changeset_ids`
-- column on `campaigns`

CREATE OR REPLACE FUNCTION delete_changeset_reference_on_campaigns()
    RETURNS TRIGGER AS
$delete_changeset_reference_on_campaigns$
    BEGIN
        UPDATE
          campaigns
        SET
          changeset_ids = campaigns.changeset_ids - OLD.id::text
        WHERE
          campaigns.changeset_ids ? OLD.id::text;

        RETURN OLD;
    END;
$delete_changeset_reference_on_campaigns$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trig_delete_changeset_reference_on_campaigns ON changesets;

CREATE TRIGGER trig_delete_changeset_reference_on_campaigns
AFTER DELETE on changesets
FOR EACH ROW EXECUTE PROCEDURE delete_changeset_reference_on_campaigns();

-- The reverse:
-- When we delete a `campaign` we remove its ID from the `campaign_ids`
-- column on `changesets`

CREATE OR REPLACE FUNCTION delete_campaign_reference_on_changesets()
    RETURNS TRIGGER AS
$delete_campaign_reference_on_changesets$
    BEGIN
        UPDATE
          changesets
        SET
          campaign_ids = changesets.campaign_ids - OLD.id::text
        WHERE
          changesets.campaign_ids ? OLD.id::text;

        RETURN OLD;
    END;
$delete_campaign_reference_on_changesets$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trig_delete_campaign_reference_on_changesets ON campaigns;

CREATE TRIGGER trig_delete_campaign_reference_on_changesets
AFTER DELETE on campaigns
FOR EACH ROW EXECUTE PROCEDURE delete_campaign_reference_on_changesets();

COMMIT;
