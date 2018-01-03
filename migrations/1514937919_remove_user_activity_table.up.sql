BEGIN;
ALTER TABLE users ADD COLUMN page_views integer not null default 0;
ALTER TABLE users ADD COLUMN search_queries integer not null default 0;
UPDATE users SET
	page_views=(SELECT page_views FROM user_activity WHERE user_activity.user_id=users.id),
	search_queries=(SELECT search_queries FROM user_activity WHERE user_activity.user_id=users.id);
DROP TABLE user_activity;
COMMIT;
