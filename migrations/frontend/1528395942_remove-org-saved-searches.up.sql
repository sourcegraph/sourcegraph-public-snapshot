BEGIN;

DELETE FROM saved_searches WHERE user_id IS NULL;

ALTER TABLE saved_searches
	ALTER COLUMN user_id SET NOT NULL;

COMMENT ON COLUMN saved_searches.org_id IS 'DEPRECATED: saved searches must be owned by a user';

COMMIT;
