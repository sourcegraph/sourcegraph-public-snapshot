BEGIN;

ALTER TABLE user_permissions DROP COLUMN version;

COMMIT;
