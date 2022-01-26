-- +++
-- parent: 1528395942
-- +++

BEGIN;

UPDATE saved_searches
SET (notify_owner, notify_slack) = (false, false);

ALTER TABLE saved_searches
	ADD CONSTRAINT saved_searches_notifications_disabled CHECK (
		notify_owner = false
		AND notify_slack = false
	);

ALTER TABLE IF EXISTS saved_searches
	ALTER COLUMN user_id DROP NOT NULL;

COMMIT;
