BEGIN;
-- NOTE: this migration deleted saved searches belonging to orgs when what we should have done
-- was remove notifications from saved searches. Saved searches are still used (and useful) as
-- a bookmarking feature, and are not deprecated. Saved search notifications, however, are
-- deprecated, and will be removed in v3.34.0 by removing the query runner service. This
-- migration specifically deleted org-owned saved searches because there is no clear user
-- to run org-saved searches as, and we were previously running saved searches as site admin,
-- but that changed as part of the PR that introduced this migration.

-- DELETE FROM saved_searches WHERE user_id IS NULL;

-- ALTER TABLE saved_searches
-- 	ALTER COLUMN user_id SET NOT NULL;

-- COMMENT ON COLUMN saved_searches.org_id IS 'DEPRECATED: saved searches must be owned by a user';

UPDATE saved_searches
SET (notify_owner, notify_slack) = (false, false);

ALTER TABLE saved_searches
	ADD CONSTRAINT saved_searches_notifications_disabled CHECK (
		notify_owner = false
		AND notify_slack = false
	);

COMMIT;
