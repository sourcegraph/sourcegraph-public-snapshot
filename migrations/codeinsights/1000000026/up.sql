-- +++
-- parent: 1000000025
-- +++

BEGIN;

ALTER TABLE IF EXISTS commit_index
    ADD COLUMN IF NOT EXISTS indexed_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ADD COLUMN IF NOT EXISTS debug_field TEXT;

COMMIT;
