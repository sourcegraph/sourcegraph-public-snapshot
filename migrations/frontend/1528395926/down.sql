BEGIN;

CREATE INDEX sub_repo_perms_repo_id ON sub_repo_permissions (repo_id);

CREATE INDEX sub_repo_perms_user_id ON sub_repo_permissions (user_id);

CREATE INDEX sub_repo_perms_version ON sub_repo_permissions (version);

COMMIT;
