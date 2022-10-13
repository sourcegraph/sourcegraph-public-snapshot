DROP TABLE IF EXISTS insight_series_backfill;

ALTER TABLE insights_background_jobs
    DROP COLUMN IF EXISTS backfill_id;

DROP VIEW IF EXISTS insights_jobs_backfill_ready;

DROP VIEW IF EXISTS insights_jobs_backfill_new;
