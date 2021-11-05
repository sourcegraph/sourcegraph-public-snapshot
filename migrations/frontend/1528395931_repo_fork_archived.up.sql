BEGIN;

CREATE INDEX IF NOT EXISTS repo_fork_archived_idx ON repo (fork, archived) INCLUDE (id, name, private) WHERE (deleted_at IS NULL AND blocked IS NULL);

COMMIT;
