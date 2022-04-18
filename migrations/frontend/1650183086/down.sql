-- Undo the changes made in the up migration

DROP INDEX IF EXISTS github_internal_repo_user_permissions_user_id;

DROP TABLE IF EXISTS github_internal_repo_user_permissions;
