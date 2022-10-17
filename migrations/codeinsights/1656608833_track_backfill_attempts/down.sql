-- Undo the changes made in the up migration
ALTER TABLE IF EXISTS insight_series DROP COLUMN IF EXISTS backfill_attempts;
