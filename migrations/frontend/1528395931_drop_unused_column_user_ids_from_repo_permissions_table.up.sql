-- +++
-- parent: 1528395930
-- +++

BEGIN;

ALTER TABLE IF EXISTS repo_permissions DROP COLUMN IF EXISTS user_ids;

COMMIT;
