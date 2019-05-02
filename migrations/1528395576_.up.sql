BEGIN;

-- This trigger was part of a backcompat migration from a long time ago. It
-- was from when we renamed the URI column to Name column. It is no longer
-- needed.
ALTER TABLE repo ALTER COLUMN uri DROP NOT NULL;
DROP TRIGGER IF EXISTS trig_set_repo_name ON repo;
DROP FUNCTION IF EXISTS set_repo_name();

COMMIT;
