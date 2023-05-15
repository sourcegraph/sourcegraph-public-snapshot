ALTER TABLE IF EXISTS insights_data_retention_jobs
ADD COLUMN IF NOT EXISTS series_id_string text not null default '';

ALTER TABLE insights_data_retention_jobs DROP COLUMN series_id_string;
