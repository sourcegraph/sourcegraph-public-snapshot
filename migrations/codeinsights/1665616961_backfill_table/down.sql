DROP VIEW IF EXISTS insights_jobs_backfill_in_progress;

DROP VIEW IF EXISTS insights_jobs_backfill_new;

ALTER TABLE insights_background_jobs
    DROP COLUMN IF EXISTS backfill_id;

DROP TABLE IF EXISTS insight_series_backfill;
