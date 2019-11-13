BEGIN;

UPDATE lsif_commits SET parent_commit = '' where parent_commit IS NULL;
ALTER TABLE lsif_commits ALTER COLUMN parent_commit SET NOT NULL;

COMMIT;
