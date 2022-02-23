ALTER TABLE IF EXISTS commit_index
    DROP COLUMN IF EXISTS indexed_at,
    DROP COLUMN IF EXISTS debug_field;
