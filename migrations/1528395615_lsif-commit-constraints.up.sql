-- Note: `commit` is a reserved word, so it's quoted.

BEGIN;

ALTER TABLE lsif_dumps ADD CONSTRAINT lsif_dumps_commit_valid_chars CHECK ("commit" ~ '^[a-z0-9]{40}$');
ALTER TABLE lsif_commits ADD CONSTRAINT lsif_commits_commit_valid_chars CHECK ("commit" ~ '^[a-z0-9]{40}$');
ALTER TABLE lsif_commits ADD CONSTRAINT lsif_commits_parent_commit_valid_chars CHECK (parent_commit ~ '^[a-z0-9]{40}$');
ALTER TABLE lsif_dumps DROP CONSTRAINT lsif_dumps_commit_check;

COMMIT;
