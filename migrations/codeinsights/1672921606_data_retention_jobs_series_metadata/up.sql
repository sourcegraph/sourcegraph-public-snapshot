ALTER TABLE IF EXISTS insights_data_retention_jobs
ADD COLUMN series_id_string text not null default '';
