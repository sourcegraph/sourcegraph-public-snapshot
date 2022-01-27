BEGIN;

ALTER TABLE IF EXISTS saved_searches
	DROP CONSTRAINT IF EXISTS saved_searches_notifications_disabled;

COMMIT;
