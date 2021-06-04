BEGIN;

ALTER TABLE out_of_band_migrations DROP COLUMN is_enterprise;

COMMIT;
