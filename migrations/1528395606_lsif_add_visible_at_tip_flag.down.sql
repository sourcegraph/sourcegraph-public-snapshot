-- Note: `commit` is a reserved word, so it's quoted.

BEGIN;

ALTER TABLE lsif_dumps DROP COLUMN visible_at_tip;
CREATE INDEX lsif_commits_commit on lsif_commits(repository, "commit");
CREATE INDEX lsif_commits_repo_commit on lsif_commits(repository, "commit");
CREATE INDEX lsif_commits_repo_parent_commit on lsif_commits(repository, parent_commit);

COMMIT;
