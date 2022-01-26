-- +++
-- parent: 1528395929
-- +++

BEGIN;

ALTER TABLE IF EXISTS user_permissions DROP COLUMN IF EXISTS object_ids;

COMMIT;
