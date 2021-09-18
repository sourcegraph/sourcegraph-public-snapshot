BEGIN;

-- No out of band migrations should be marked as deprecated yet
-- as we don't have any utilities in place to check whether or
-- not migrations have yet completed.

UPDATE out_of_band_migrations SET deprecated = NULL;

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
