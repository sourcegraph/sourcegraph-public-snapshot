BEGIN;

ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS changeset_ids jsonb NOT NULL DEFAULT '{}'::jsonb CHECK (jsonb_typeof(changeset_ids) = 'object'::text);
CREATE INDEX IF NOT EXISTS campaigns_changeset_ids_gin_idx ON campaigns USING GIN (changeset_ids jsonb_ops);

WITH changesets AS (
    SELECT id, campaign_ids FROM changesets
)
UPDATE campaigns SET changeset_ids = changeset_ids || jsonb_build_object(changesets.id::TEXT, NULL) FROM changesets WHERE changesets.campaign_ids ? campaigns.id::TEXT;

CREATE FUNCTION delete_changeset_reference_on_campaigns() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
    BEGIN
        UPDATE
          campaigns
        SET
          changeset_ids = campaigns.changeset_ids - OLD.id::text
        WHERE
          campaigns.changeset_ids ? OLD.id::text;

        RETURN OLD;
    END;
$$;

CREATE TRIGGER trig_delete_changeset_reference_on_campaigns AFTER DELETE ON changesets FOR EACH ROW EXECUTE PROCEDURE delete_changeset_reference_on_campaigns();

COMMIT;
