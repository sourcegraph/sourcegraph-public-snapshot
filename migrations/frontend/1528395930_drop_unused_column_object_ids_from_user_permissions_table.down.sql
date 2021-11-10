BEGIN;

ALTER TABLE IF EXISTS user_permissions ADD COLUMN object_ids bytea NOT NULL default '\x';

COMMIT;
