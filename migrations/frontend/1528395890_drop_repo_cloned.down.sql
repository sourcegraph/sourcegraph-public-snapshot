BEGIN;

-- This migration is destructive since the column has been
-- deprecated since 3.26
ALTER TABLE
    repo
ADD COLUMN IF NOT EXISTS
    cloned bool NOT NULL DEFAULT false;

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
