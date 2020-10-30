BEGIN;

ALTER TABLE changeset_specs ADD COLUMN IF NOT EXISTS type TEXT;
UPDATE changeset_specs SET type = 'existing' WHERE (changeset_specs.spec->'externalID')::text IS NOT NULL;
UPDATE changeset_specs SET type = 'branch'   WHERE (changeset_specs.spec->'externalID')::text IS NULL;
ALTER TABLE changeset_specs ALTER COLUMN type SET NOT NULL;

ALTER TABLE changeset_specs ADD COLUMN IF NOT EXISTS external_id TEXT;
UPDATE changeset_specs SET external_id = changeset_specs.spec->>'externalID' WHERE changeset_specs.spec->>'externalID' IS NOT NULL;

ALTER TABLE changeset_specs ADD COLUMN IF NOT EXISTS head_ref TEXT;
UPDATE changeset_specs SET head_ref = changeset_specs.spec->>'headRef' WHERE changeset_specs.spec->>'headRef' IS NOT NULL;

UPDATE changesets SET external_branch = CONCAT('refs/heads/', external_branch) WHERE external_branch IS NOT NULL AND external_branch NOT LIKE 'refs/heads/%';

COMMIT;
