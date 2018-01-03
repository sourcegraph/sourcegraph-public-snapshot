BEGIN;
ALTER TABLE users ADD COLUMN page_views integer default 0;
ALTER TABLE users ADD COLUMN search_queries integer default 0;
UPDATE users SET
	page_views=(SELECT page_views FROM user_activity WHERE user_activity.user_id=users.id),
	search_queries=(SELECT search_queries FROM user_activity WHERE user_activity.user_id=users.id);
UPDATE users SET page_views=0 WHERE page_views IS NULL;
UPDATE users SET search_queries=0 WHERE search_queries IS NULL;
ALTER TABLE users ALTER COLUMN page_views SET NOT NULL;
ALTER TABLE users ALTER COLUMN search_queries SET NOT NULL;
DROP TABLE user_activity;
COMMIT;
