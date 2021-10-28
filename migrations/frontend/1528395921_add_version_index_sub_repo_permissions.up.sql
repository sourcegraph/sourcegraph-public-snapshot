BEGIN;

drop index if exists sub_repo_permissions_repo_id_user_id_uindex;

create unique index sub_repo_permissions_repo_id_user_id_version_uindex
    on sub_repo_permissions (repo_id, user_id, version);

create index sub_repo_perms_version ON sub_repo_permissions (version);

COMMIT;
