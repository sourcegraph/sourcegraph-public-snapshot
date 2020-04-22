-- Note: `commit` is a reserved word, so it's quoted.

BEGIN;

--
-- lsif_uploads
--

-- Restore old column
ALTER TABLE lsif_uploads RENAME COLUMN repository_name_at_upload TO repository;
CREATE UNIQUE INDEX lsif_uploads_repository_commit_root on lsif_uploads(repository, "commit", root) WHERE state = 'completed'::lsif_upload_state;
CREATE INDEX lsif_uploads_visible_repository_commit on lsif_uploads(repository, "commit") WHERE visible_at_tip;
ALTER TABLE lsif_uploads ADD CONSTRAINT "lsif_uploads_repository_check" CHECK (repository <> ''::text);

-- Re-populate old column
UPDATE lsif_uploads u SET repository = (SELECT name FROM repo r WHERE r.id = u.repository_id LIMIT 1);
ALTER TABLE lsif_uploads ALTER COLUMN repository_id SET NOT NULL;

-- Drop view dependent on new column
DROP VIEW lsif_dumps;

-- Drop new column
DROP INDEX lsif_uploads_repository_id_commit_root;
DROP INDEX lsif_uploads_visible_repository_id_commit;
ALTER TABLE lsif_uploads DROP repository_id;

-- Recreate view with new column names
CREATE VIEW lsif_dumps AS SELECT u.*, u.finished_at as processed_at FROM lsif_uploads u WHERE state = 'completed';

--
-- lsif_commits
--

-- Restore old column
ALTER TABLE lsif_commits ADD repository text;
CREATE UNIQUE INDEX lsif_commits_repo_commit_parent_commit_unique ON lsif_commits(repository, "commit", parent_commit);
CREATE INDEX lsif_commits_parent_commit ON lsif_commits(repository, parent_commit);

-- Re-populate old column
UPDATE lsif_commits c SET repository = (SELECT name FROM repo r WHERE r.id = c.repository_id LIMIT 1);
ALTER TABLE lsif_commits ALTER COLUMN repository_id SET NOT NULL;

-- Drop new column
DROP INDEX lsif_commits_repository_id_commit_parent_commit_unique;
DROP INDEX lsif_commits_repository_id_parent_commit;
ALTER TABLE lsif_commits DROP repository_id;

COMMIT;
