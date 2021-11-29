BEGIN;

ALTER TABLE IF EXISTS repo_permissions ADD COLUMN user_ids bytea NOT NULL default '\x';

COMMIT;
