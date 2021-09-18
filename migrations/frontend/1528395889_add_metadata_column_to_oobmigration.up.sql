BEGIN;

ALTER TABLE
    out_of_band_migrations
ADD COLUMN IF NOT EXISTS
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb;

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
