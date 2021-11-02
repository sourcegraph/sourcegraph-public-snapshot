BEGIN;

CREATE INDEX sub_repo_perms_repo_id ON sub_repo_permissions (repo_id);

COMMIT;
