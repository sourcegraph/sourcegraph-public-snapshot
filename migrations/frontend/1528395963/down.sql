DROP TABLE IF EXISTS insights_settings_migration_jobs;

DELETE
FROM out_of_band_migrations
WHERE id = 14;
