TRUNCATE TABLE user_pending_permissions;
TRUNCATE TABLE repo_pending_permissions;

ALTER TABLE IF EXISTS user_pending_permissions ALTER COLUMN id TYPE int;
ALTER TABLE IF EXISTS repo_pending_permissions ALTER COLUMN user_ids_ints TYPE int[];

ALTER SEQUENCE IF EXISTS user_pending_permissions_id_seq RESTART;
