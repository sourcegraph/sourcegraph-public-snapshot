BEGIN;

TRUNCATE TABLE repo_permissions;
TRUNCATE TABLE user_permissions;
TRUNCATE TABLE repo_pending_permissions;
TRUNCATE TABLE user_pending_permissions;

ALTER TABLE repo_permissions DROP COLUMN IF EXISTS user_ids;
ALTER TABLE user_permissions DROP COLUMN IF EXISTS object_ids;
ALTER TABLE repo_pending_permissions DROP COLUMN IF EXISTS user_ids;
ALTER TABLE user_pending_permissions DROP COLUMN IF EXISTS object_ids;

CREATE EXTENSION IF NOT EXISTS intarray;

ALTER TABLE repo_permissions ADD COLUMN user_ids INT[] NOT NULL DEFAULT '{}';
ALTER TABLE user_permissions ADD COLUMN object_ids INT[] NOT NULL DEFAULT '{}';
ALTER TABLE repo_pending_permissions ADD COLUMN user_ids INT[] NOT NULL DEFAULT '{}';
ALTER TABLE user_pending_permissions ADD COLUMN object_ids INT[] NOT NULL DEFAULT '{}';

COMMIT;
