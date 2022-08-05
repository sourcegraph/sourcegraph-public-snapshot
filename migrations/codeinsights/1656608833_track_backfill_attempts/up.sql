-- Perform migration here.
--
ALTER TABLE IF EXISTS insight_series
    ADD COLUMN IF NOT EXISTS backfill_attempts INT NOT NULL DEFAULT 0;
