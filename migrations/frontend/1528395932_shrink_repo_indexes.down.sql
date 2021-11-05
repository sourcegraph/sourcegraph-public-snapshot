BEGIN;

CREATE INDEX IF NOT EXISTS repo_archived ON repo (archived);
DROP INDEX IF EXISTS repo_archived_idx;

CREATE INDEX IF NOT EXISTS repo_blocked_idx ON repo ((blocked IS NOT NULL));

CREATE INDEX IF NOT EXISTS repo_created_at ON repo (created_at);
DROP INDEX IF EXISTS repo_created_at_idx;

CREATE INDEX IF NOT EXISTS repo_fork ON repo (fork);
DROP INDEX IF EXISTS repo_fork_idx;

CREATE INDEX IF NOT EXISTS repo_is_not_blocked_idx ON repo ((blocked IS NULL));

CREATE INDEX IF NOT EXISTS repo_non_deleted_id_name_idx ON repo (id, name) WHERE deleted_at IS NULL;
DROP INDEX IF EXISTS repo_id_idx;

CREATE INDEX IF NOT EXISTS repo_private ON repo (private);
DROP INDEX IF EXISTS repo_private_idx;

CREATE INDEX IF NOT EXISTS repo_name_idx ON repo (lower(name::text) COLLATE "C");
DROP INDEX IF EXISTS repo_name_idx;

CREATE INDEX IF NOT EXISTS repo_name_trgm ON repo USING gin (lower(name::text) gin_trgm_ops);
DROP INDEX IF EXISTS repo_name_trgm;

COMMIT;
