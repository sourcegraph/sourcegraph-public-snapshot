-- +++
-- parent: 1528395925
-- +++

BEGIN;

DROP INDEX IF EXISTS sub_repo_perms_repo_id;

DROP INDEX IF EXISTS sub_repo_perms_user_id;

DROP INDEX IF EXISTS sub_repo_perms_version;

COMMIT;
