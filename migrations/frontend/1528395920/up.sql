BEGIN;

create table sub_repo_permissions
(
    repo_id       integer       not null
        constraint sub_repo_permissions_repo_id_fk
            references repo
            on delete cascade,
    user_id       integer       not null
        constraint sub_repo_permissions_users_id_fk
            references users
            on delete cascade,
    version       int default 1 not null,
    path_includes text[],
    path_excludes text[],
    updated_at timestamp with time zone default now() not null
);

comment on table sub_repo_permissions is 'Responsible for storing permissions at a finer granularity than repo';

create unique index sub_repo_permissions_repo_id_user_id_uindex
    on sub_repo_permissions (repo_id, user_id);

create index sub_repo_perms_user_id ON sub_repo_permissions (user_id);
create index sub_repo_perms_repo_id ON sub_repo_permissions (repo_id);

COMMIT;
