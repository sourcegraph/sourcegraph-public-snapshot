ALTER TABLE IF EXISTS insight_series
	ADD COLUMN IF NOT EXISTS supports_augmentation BOOLEAN DEFAULT FALSE;