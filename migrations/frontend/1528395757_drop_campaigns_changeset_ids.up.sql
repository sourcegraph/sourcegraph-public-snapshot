BEGIN;

DROP TRIGGER IF EXISTS trig_delete_changeset_reference_on_campaigns ON changesets;
DROP FUNCTION IF EXISTS delete_changeset_reference_on_campaigns();

ALTER TABLE campaigns DROP COLUMN IF EXISTS changeset_ids;

COMMIT;
