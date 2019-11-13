BEGIN;

ALTER TABLE lsif_commits ALTER COLUMN parent_commit DROP NOT NULL;
UPDATE lsif_commits SET parent_commit = NULL where parent_commit = '';

COMMIT;
