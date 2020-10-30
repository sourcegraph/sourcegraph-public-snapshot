BEGIN;

ALTER TABLE changeset_specs DROP COLUMN IF EXISTS type;
ALTER TABLE changeset_specs DROP COLUMN IF EXISTS external_id;
ALTER TABLE changeset_specs DROP COLUMN IF EXISTS head_ref;

UPDATE changesets SET external_branch = LTRIM(external_branch, 'refs/heads/') WHERE external_branch IS NOT NULL;

COMMIT;
