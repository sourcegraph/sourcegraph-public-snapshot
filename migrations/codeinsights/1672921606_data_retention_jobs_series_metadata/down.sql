ALTER TABLE IF EXISTS insight_data_retention_jobs
DROP COLUMN IF EXISTS series_id_string;

ALTER TABLE IF EXISTS insights_data_retention_jobs DROP COLUMN IF EXISTS series_id_string;
