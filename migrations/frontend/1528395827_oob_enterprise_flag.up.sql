BEGIN;

-- Add enterprise flag to out of band migrations so we know what to ignore in OSS
-- If we don't run the migrators but the progress for the enterprise migration
-- record stays at 0% it will block all upgrades after the migration deprecation.

ALTER TABLE out_of_band_migrations ADD COLUMN is_enterprise boolean DEFAULT false;
UPDATE out_of_band_migrations SET is_enterprise = true WHERE id NOT IN (3, 6);
ALTER TABLE out_of_band_migrations ALTER COLUMN is_enterprise SET NOT NULL;

COMMENT ON COLUMN out_of_band_migrations.is_enterprise IS 'When true, these migrations are invisible to OSS mode.';

COMMIT;
