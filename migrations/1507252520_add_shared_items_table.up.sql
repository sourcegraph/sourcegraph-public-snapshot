CREATE TABLE shared_items (
	id bigserial NOT NULL PRIMARY KEY,
	ulid text NOT NULL,
	author_user_id text NOT NULL,
	thread_id bigint,
	comment_id bigint,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
	updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
	deleted_at TIMESTAMP WITH TIME ZONE
);
CREATE UNIQUE INDEX shared_items_ulid_idx ON shared_items USING btree(ulid);
