BEGIN;

ALTER TABLE changesets DROP CONSTRAINT IF EXISTS external_branch_ref_prefix;

UPDATE changesets SET external_branch = LTRIM(external_branch, 'refs/heads/') WHERE external_branch IS NOT NULL;

UPDATE changeset_specs SET spec = jsonb_set(spec, '{headRef}', to_jsonb(LTRIM(spec->>'headRef', 'refs/heads/'))) WHERE spec->>'headRef' IS NOT NULL AND spec->>'headRef' != '' AND spec->>'headRef' LIKE 'refs/heads/%';
UPDATE changeset_specs SET spec = jsonb_set(spec, '{baseRef}', to_jsonb(LTRIM(spec->>'baseRef', 'refs/heads/'))) WHERE spec->>'baseRef' IS NOT NULL AND spec->>'baseRef' != '' AND spec->>'baseRef' LIKE 'refs/heads/%';

COMMIT;
