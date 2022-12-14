ALTER TABLE IF EXISTS insight_series
	ADD COLUMN IF NOT EXISTS repository_criteria text;

COMMENT ON COLUMN insight_series.repository_criteria IS 'The search criteria used to determine the repositories that are included in this series.';

