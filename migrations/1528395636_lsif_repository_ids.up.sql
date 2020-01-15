-- Note: `commit` is a reserved word, so it's quoted.

BEGIN;

--
-- lsif_uploads
--

-- Rename column
ALTER TABLE lsif_uploads RENAME COLUMN repository TO repository_name_at_upload;

-- Add new repository identifier column
ALTER TABLE lsif_uploads ADD repository_id int;
CREATE UNIQUE INDEX lsif_uploads_repository_id_commit_root ON lsif_uploads(repository_id, "commit", root) WHERE state = 'completed'::lsif_upload_state;
CREATE INDEX lsif_uploads_visible_repository_id_commit ON lsif_uploads(repository_id, "commit") WHERE visible_at_tip;

-- Populate new column and delete any uploads that we can't correlate
UPDATE lsif_uploads u SET repository_id = (SELECT id FROM repo r WHERE r.name = u.repository_name_at_upload LIMIT 1);
DELETE FROM lsif_uploads WHERE repository_id IS NULL;
ALTER TABLE lsif_uploads ALTER COLUMN repository_id SET NOT NULL;

-- Drop view dependent on old column
DROP VIEW lsif_dumps;

-- Drop old column constraints/indexes
DROP INDEX lsif_uploads_repository_commit_root;
DROP INDEX lsif_uploads_visible_repository_commit;
ALTER TABLE lsif_uploads DROP CONSTRAINT lsif_uploads_repository_check;

-- Recreate view with new column names
CREATE VIEW lsif_dumps AS SELECT u.*, u.finished_at as processed_at FROM lsif_uploads u WHERE state = 'completed';

--
-- lsif_commits
--

-- Add new column
ALTER TABLE lsif_commits ADD repository_id int;
CREATE UNIQUE INDEX lsif_commits_repository_id_commit_parent_commit_unique ON lsif_commits(repository_id, "commit", parent_commit);
CREATE INDEX lsif_commits_repository_id_parent_commit ON lsif_commits(repository_id, parent_commit);

-- Populate new column and delete any commits that we can't correlate
UPDATE lsif_commits c SET repository_id = (SELECT id FROM repo r WHERE r.name = c.repository LIMIT 1);
DELETE FROM lsif_commits WHERE repository_id IS NULL;
ALTER TABLE lsif_commits ALTER COLUMN repository_id SET NOT NULL;

-- Drop old column
DROP INDEX lsif_commits_repo_commit_parent_commit_unique;
DROP INDEX lsif_commits_parent_commit;
ALTER TABLE lsif_commits DROP repository;

COMMIT;
