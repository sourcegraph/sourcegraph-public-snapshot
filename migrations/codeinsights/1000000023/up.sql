ALTER TABLE IF EXISTS insight_series
    ADD COLUMN IF NOT EXISTS generated_from_capture_groups BOOL NOT NULL DEFAULT FALSE;
