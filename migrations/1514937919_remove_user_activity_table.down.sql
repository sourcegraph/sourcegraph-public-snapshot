CREATE TABLE user_activity (
	id serial NOT NULL PRIMARY KEY,
	user_id integer NOT NULL UNIQUE,
	page_views integer NOT NULL DEFAULT 0,
	search_queries integer NOT NULL DEFAULT 0,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
	updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
	CONSTRAINT user_activity FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE RESTRICT

);
-- Preserving user_activity created_at and updated_at values is not necessary.
INSERT INTO user_activity(user_id, page_views, search_queries) SELECT id AS user_id, page_views, search_queries FROM users;
ALTER TABLE users DROP COLUMN page_views;
ALTER TABLE users DROP COLUMN search_queries;
