BEGIN;

ALTER TABLE
    out_of_band_migrations
DROP COLUMN IF EXISTS metadata;

COMMIT;
