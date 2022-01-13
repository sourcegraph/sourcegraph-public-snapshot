-- +++
-- parent: 1528395888
-- +++

BEGIN;

ALTER TABLE
    out_of_band_migrations
ADD COLUMN IF NOT EXISTS
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb;

COMMIT;
