BEGIN;

-- We don't need to reconstruct the contents of the raw_spec column (and,
-- indeed, we can't), so we'll just leave it with empty strings.

ALTER TABLE
    changeset_specs
ADD COLUMN IF NOT EXISTS
    raw_spec TEXT NOT NULL DEFAULT '';

COMMIT;
