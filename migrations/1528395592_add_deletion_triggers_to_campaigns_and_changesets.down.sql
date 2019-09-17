BEGIN;

DROP TRIGGER IF EXISTS trig_delete_changeset_reference_on_campaigns ON changesets;
DROP FUNCTION IF EXISTS delete_changeset_reference_on_campaigns();

DROP TRIGGER IF EXISTS trig_delete_campaign_reference_on_changesets ON campaigns;
DROP FUNCTION IF EXISTS delete_campaign_reference_on_changesets();

COMMIT;
