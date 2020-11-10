BEGIN;

UPDATE changesets SET external_branch = CONCAT('refs/heads/', external_branch) WHERE external_branch IS NOT NULL AND external_branch NOT LIKE 'refs/heads/%';

COMMIT;
