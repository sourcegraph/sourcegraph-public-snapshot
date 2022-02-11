BEGIN;

-- We changed the default timeframe that the historical insights builder produces from:
--
-- 1 data point per week, for the last 52 weeks
--
-- To:
--
-- 1 data point per month, for the last 6 months
--
-- To avoid any confusion and just start fresh, we wipe all data here for now. This isn't
-- needed in general when making this change, but is useful in this specific situation.
DELETE FROM series_points;

COMMIT;
