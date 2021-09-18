BEGIN;

ALTER TABLE out_of_band_migrations DROP COLUMN is_enterprise;

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
