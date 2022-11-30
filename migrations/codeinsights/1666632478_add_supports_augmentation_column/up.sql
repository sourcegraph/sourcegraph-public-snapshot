ALTER TABLE IF EXISTS insight_series
	ADD COLUMN IF NOT EXISTS supports_augmentation BOOLEAN NOT NULL DEFAULT TRUE;

UPDATE insight_series SET supports_augmentation = FALSE;