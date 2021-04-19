BEGIN;

-- No out of band migrations should be marked as deprecated yet
-- as we don't have any utilities in place to check whether or
-- not migrations have yet completed.

UPDATE out_of_band_migrations SET deprecated = NULL;

COMMIT;
