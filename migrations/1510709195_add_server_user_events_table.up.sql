CREATE TABLE user_activity (
	id serial NOT NULL PRIMARY KEY,
	user_id integer NOT NULL UNIQUE,
	page_views integer NOT NULL DEFAULT 0,
	search_queries integer NOT NULL DEFAULT 0,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
	updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
	CONSTRAINT user_activity FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE RESTRICT
);
