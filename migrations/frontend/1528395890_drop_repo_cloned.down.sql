BEGIN;

-- This migration is destructive since the column has been
-- deprecated since 3.26
ALTER TABLE
    repo
ADD COLUMN IF NOT EXISTS
    cloned bool NOT NULL DEFAULT false;

COMMIT;
