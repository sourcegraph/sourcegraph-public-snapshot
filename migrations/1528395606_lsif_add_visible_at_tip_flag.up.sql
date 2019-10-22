BEGIN;

ALTER TABLE lsif_dumps ADD COLUMN visible_at_tip boolean NOT NULL DEFAULT false;

-- These are duplicate indexes
DROP INDEX IF EXISTS lsif_commits_commit;
DROP INDEX IF EXISTS lsif_commits_repo_commit;
DROP INDEX IF EXISTS lsif_commits_repo_parent_commit;

COMMIT;
