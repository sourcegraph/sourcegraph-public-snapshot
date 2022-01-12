-- +++
-- parent: 1528395941
-- +++

BEGIN;
-- NOTE: this migration deleted saved searches belonging to orgs when what we should have done
-- was remove notifications from saved searches. Saved searches are still used (and useful) as
-- a bookmarking feature, and are not deprecated. Saved search notifications, however, are
-- deprecated, and will be removed in v3.34.0 by removing the query runner service. 

-- DELETE FROM saved_searches WHERE user_id IS NULL;

-- ALTER TABLE saved_searches
-- 	ALTER COLUMN user_id SET NOT NULL;

-- COMMENT ON COLUMN saved_searches.org_id IS 'DEPRECATED: saved searches must be owned by a user';

COMMIT;
