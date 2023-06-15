-- take no action on down so that the next up will not trigger another refreshing of series data.
ALTER TABLE insight_series DROP COLUMN IF EXISTS needs_migration;
