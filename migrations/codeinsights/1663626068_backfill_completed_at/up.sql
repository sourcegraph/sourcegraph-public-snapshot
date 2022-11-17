ALTER TABLE IF EXISTS insight_series
    ADD COLUMN IF NOT EXISTS backfill_completed_at TIMESTAMP;
