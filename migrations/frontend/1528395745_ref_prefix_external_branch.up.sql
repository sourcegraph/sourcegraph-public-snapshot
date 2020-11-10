BEGIN;

UPDATE changesets SET external_branch = CONCAT('refs/heads/', external_branch) WHERE external_branch IS NOT NULL AND external_branch NOT LIKE 'refs/heads/%';

UPDATE changeset_specs SET spec = jsonb_set(spec, '{headRef}', to_jsonb(CONCAT('refs/heads/', spec->>'headRef'))) WHERE spec->>'headRef' IS NOT NULL AND spec->>'headRef' != '' AND spec->>'headRef' NOT LIKE 'refs/heads/%';
UPDATE changeset_specs SET spec = jsonb_set(spec, '{baseRef}', to_jsonb(CONCAT('refs/heads/', spec->>'baseRef'))) WHERE spec->>'baseRef' IS NOT NULL AND spec->>'baseRef' != '' AND spec->>'baseRef' NOT LIKE 'refs/heads/%';

ALTER TABLE changesets ADD CONSTRAINT external_branch_ref_prefix CHECK (external_branch LIKE 'refs/heads/%');

COMMIT;
