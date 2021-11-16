BEGIN;

DROP TABLE IF EXISTS insights_settings_migration_jobs;

DELETE
FROM sg.public.out_of_band_migrations
WHERE id = 14;

COMMIT;
