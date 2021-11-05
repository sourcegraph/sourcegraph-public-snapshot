BEGIN;

DROP INDEX IF EXISTS repo_archived;
CREATE INDEX IF NOT EXISTS repo_archived_idx ON repo (archived) INCLUDE (id, name, private) WHERE (deleted_at IS NULL AND blocked IS NULL);

DROP INDEX IF EXISTS repo_blocked_idx;

DROP INDEX IF EXISTS repo_created_at;
CREATE INDEX IF NOT EXISTS repo_created_at_idx ON repo (created_at) INCLUDE (id, name, private) WHERE (deleted_at IS NULL AND blocked IS NULL);

DROP INDEX IF EXISTS repo_fork;
CREATE INDEX IF NOT EXISTS repo_fork_idx ON repo (fork) INCLUDE (id, name, private) WHERE (deleted_at IS NULL AND blocked IS NULL);

DROP INDEX IF EXISTS repo_is_not_blocked_idx;

DROP INDEX IF EXISTS repo_non_deleted_id_name_idx;
CREATE INDEX IF NOT EXISTS repo_id_idx ON repo (id) INCLUDE (name, private) WHERE deleted_at IS NULL AND blocked IS NULL;

DROP INDEX IF EXISTS repo_private;
CREATE INDEX IF NOT EXISTS repo_private_idx ON repo (private) INCLUDE (id, name) WHERE (deleted_at IS NULL AND blocked IS NULL);

DROP INDEX IF EXISTS repo_name_idx;
CREATE INDEX IF NOT EXISTS repo_name_idx ON repo (lower((name)::text) COLLATE "C") INCLUDE (id, private) WHERE (deleted_at IS NULL AND blocked IS NULL);

DROP INDEX IF EXISTS repo_name_trgm;
CREATE INDEX IF NOT EXISTS repo_name_trgm ON repo USING gin (lower((name)::text) gin_trgm_ops) WHERE (deleted_at IS NULL AND blocked IS NULL);

COMMIT;
