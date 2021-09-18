BEGIN;

DROP INDEX IF EXISTS repo_is_blocked_idx;
DROP INDEX IF EXISTS repo_is_not_blocked_idx;

DROP FUNCTION IF EXISTS repo_block;

ALTER TABLE IF EXISTS repo DROP COLUMN IF EXISTS blocked;

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
