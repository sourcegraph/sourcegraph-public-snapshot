BEGIN;

-- This trigger was part of a backcompat migration from a long time ago. It
-- was from when we renamed the URI column to Name column. It is no longer
-- needed.
ALTER TABLE repo ALTER COLUMN uri DROP NOT NULL;
DROP TRIGGER IF EXISTS trig_set_repo_name ON repo;
DROP FUNCTION IF EXISTS set_repo_name();

-- These are unused columns since 3.0
-- https://github.com/sourcegraph/sourcegraph/issues/644
ALTER TABLE repo
  DROP COLUMN IF EXISTS pushed_at,
  DROP COLUMN IF EXISTS indexed_revision,
  DROP COLUMN IF EXISTS freeze_indexed_revision;

COMMIT;
