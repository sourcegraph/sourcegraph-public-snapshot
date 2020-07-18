BEGIN;

TRUNCATE TABLE repo_permissions;
TRUNCATE TABLE user_permissions;
TRUNCATE TABLE repo_pending_permissions;
TRUNCATE TABLE user_pending_permissions;

ALTER TABLE repo_permissions DROP COLUMN IF EXISTS user_ids;
ALTER TABLE user_permissions DROP COLUMN IF EXISTS object_ids;
ALTER TABLE repo_pending_permissions DROP COLUMN IF EXISTS user_ids;
ALTER TABLE user_pending_permissions DROP COLUMN IF EXISTS object_ids;

-- Reset values to be binary format of empty roaring bitmaps
ALTER TABLE user_permissions ADD COLUMN object_ids BYTEA NOT NULL DEFAULT '\x3a30000000000000';
ALTER TABLE repo_permissions ADD COLUMN user_ids BYTEA NOT NULL DEFAULT '\x3a30000000000000';
ALTER TABLE repo_pending_permissions ADD COLUMN object_ids BYTEA NOT NULL DEFAULT '\x3a30000000000000';
ALTER TABLE user_pending_permissions ADD COLUMN user_ids BYTEA NOT NULL DEFAULT '\x3a30000000000000';

DROP EXTENSION IF EXISTS intarray;

COMMIT;
