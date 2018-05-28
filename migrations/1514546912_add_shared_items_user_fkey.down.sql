ALTER TABLE shared_items RENAME COLUMN author_user_id TO author_user_id_new;
ALTER TABLE shared_items ADD COLUMN author_user_id text;
UPDATE shared_items SET author_user_id=(SELECT users.auth_id FROM users WHERE users.id=shared_items.author_user_id_new);
ALTER TABLE shared_items ALTER COLUMN author_user_id SET NOT NULL;
ALTER TABLE shared_items DROP COLUMN author_user_id_new;
INSERT INTO shared_items(id, ulid, author_user_id, thread_id, comment_id, created_at, updated_at, deleted_at, public)
	SELECT id, ulid, author_user_id_old, thread_id, comment_id, created_at, updated_at, deleted_at, public FROM shared_items_bkup_1514546912;
DROP TABLE shared_items_bkup_1514546912;
