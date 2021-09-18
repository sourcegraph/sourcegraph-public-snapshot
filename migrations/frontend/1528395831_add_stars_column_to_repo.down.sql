BEGIN;

DROP INDEX IF EXISTS repo_stars_idx;

ALTER TABLE repo DROP COLUMN IF EXISTS stars;

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
