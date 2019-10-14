BEGIN;

DROP VIEW IF EXISTS lsif_commits_with_lsif_data;

-- now there can be multiple LSIF dumps per repository@commit.
ALTER TABLE lsif_dumps DROP CONSTRAINT lsif_dumps_repository_commit_key;
ALTER TABLE lsif_dumps ADD COLUMN root TEXT NOT NULL DEFAULT '';

END;
