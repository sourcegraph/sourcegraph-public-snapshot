-- Note: `commit` is a reserved word, so it's quoted.

BEGIN;

ALTER TABLE lsif_dumps ADD CONSTRAINT lsif_dumps_commit_check CHECK (LENGTH("commit") = 40);
ALTER TABLE lsif_dumps DROP CONSTRAINT lsif_dumps_commit_valid_chars;
ALTER TABLE lsif_commits DROP CONSTRAINT lsif_commits_commit_valid_chars;
ALTER TABLE lsif_commits DROP CONSTRAINT lsif_commits_parent_commit_valid_chars;

COMMIT;
