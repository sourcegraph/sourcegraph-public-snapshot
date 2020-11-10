BEGIN;

UPDATE changesets SET external_branch = LTRIM(external_branch, 'refs/heads/') WHERE external_branch IS NOT NULL;

COMMIT;
