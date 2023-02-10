ALTER TABLE IF EXISTS insight_view
	ADD COLUMN IF NOT EXISTS series_num_samples INT; 