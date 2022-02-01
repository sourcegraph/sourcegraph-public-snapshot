BEGIN;

drop index if exists sub_repo_permissions_repo_id_user_id_version_uindex;
drop index if exists sub_repo_perms_version;

create unique index sub_repo_permissions_repo_id_user_id_uindex
    on sub_repo_permissions (repo_id, user_id);

COMMIT;
