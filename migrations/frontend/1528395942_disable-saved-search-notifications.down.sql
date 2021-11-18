BEGIN;

-- ALTER TABLE saved_searches
-- 	ALTER COLUMN user_id DROP NOT NULL;

ALTER TABLE IF EXISTS saved_searches
	DROP CONSTRAINT IF EXISTS saved_searches_notifications_disabled;

COMMIT;
