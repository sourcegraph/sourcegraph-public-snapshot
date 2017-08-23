# Table "public.comments"
```
    Column    |           Type           |                       Modifiers                       
--------------+--------------------------+-------------------------------------------------------
 id           | bigint                   | not null default nextval('comments_id_seq'::regclass)
 thread_id    | bigint                   | 
 contents     | text                     | 
 created_at   | timestamp with time zone | 
 updated_at   | timestamp with time zone | 
 deleted_at   | timestamp with time zone | 
 author_name  | text                     | 
 author_email | text                     | 
Indexes:
    "comments_pkey" PRIMARY KEY, btree (id)

```

# Table "public.global_dep"
```
  Column  |  Type   | Modifiers 
----------+---------+-----------
 language | text    | not null
 dep_data | jsonb   | not null
 repo_id  | integer | not null
 hints    | jsonb   | 
Indexes:
    "global_dep_idxgin" gin (dep_data jsonb_path_ops)
    "global_dep_language" btree (language)
    "global_dep_repo_id" btree (repo_id)

```

# Table "public.global_dep_private"
```
  Column  |  Type   | Modifiers 
----------+---------+-----------
 language | text    | not null
 dep_data | jsonb   | not null
 repo_id  | integer | not null
 hints    | jsonb   | 
Indexes:
    "global_dep_private_idxgin" gin (dep_data jsonb_path_ops)
    "global_dep_private_language" btree (language)
    "global_dep_private_repo_id" btree (repo_id)

```

# Table "public.local_repos"
```
    Column    |           Type           |                        Modifiers                         
--------------+--------------------------+----------------------------------------------------------
 id           | bigint                   | not null default nextval('local_repos_id_seq'::regclass)
 remote_uri   | citext                   | 
 access_token | text                     | 
 created_at   | timestamp with time zone | 
 updated_at   | timestamp with time zone | 
 deleted_at   | timestamp with time zone | 
Indexes:
    "local_repos_pkey" PRIMARY KEY, btree (id)
    "local_repos_remote_uri_access_token_key" UNIQUE CONSTRAINT, btree (remote_uri, access_token)
    "local_repos_remote_uri_idx" btree (remote_uri)

```

# Table "public.pkgs"
```
  Column  |  Type   | Modifiers 
----------+---------+-----------
 repo_id  | integer | not null
 language | text    | not null
 pkg      | jsonb   | not null
Indexes:
    "pkg_lang_idx" btree (language)
    "pkg_pkg_idx" gin (pkg jsonb_path_ops)
    "pkg_repo_idx" btree (repo_id)

```

# Table "public.repo"
```
         Column          |           Type           |                     Modifiers                     
-------------------------+--------------------------+---------------------------------------------------
 id                      | integer                  | not null default nextval('repo_id_seq'::regclass)
 uri                     | citext                   | 
 owner                   | citext                   | 
 name                    | citext                   | 
 description             | text                     | 
 vcs                     | text                     | not null
 http_clone_url          | text                     | 
 ssh_clone_url           | text                     | 
 homepage_url            | text                     | 
 default_branch          | text                     | not null
 language                | text                     | 
 blocked                 | boolean                  | 
 deprecated              | boolean                  | 
 fork                    | boolean                  | 
 mirror                  | boolean                  | 
 private                 | boolean                  | 
 created_at              | timestamp with time zone | 
 updated_at              | timestamp with time zone | 
 pushed_at               | timestamp with time zone | 
 vcs_synced_at           | timestamp with time zone | 
 indexed_revision        | text                     | 
 freeze_indexed_revision | boolean                  | 
 origin_repo_id          | text                     | 
 origin_service          | integer                  | 
 origin_api_base_url     | text                     | 
Indexes:
    "repo_pkey" PRIMARY KEY, btree (id)
    "repo_uri_unique" UNIQUE, btree (uri)
    "repo_name" btree (name text_pattern_ops)
    "repo_name_ci" btree (name)
    "repo_owner_ci" btree (owner)
    "repo_uri_trgm" gin (lower(uri::text) gin_trgm_ops)

```

# Table "public.threads"
```
     Column      |           Type           |                      Modifiers                       
-----------------+--------------------------+------------------------------------------------------
 id              | bigint                   | not null default nextval('threads_id_seq'::regclass)
 local_repo_id   | bigint                   | 
 file            | text                     | 
 revision        | text                     | 
 start_line      | integer                  | 
 end_line        | integer                  | 
 start_character | integer                  | 
 end_character   | integer                  | 
 created_at      | timestamp with time zone | 
Indexes:
    "threads_pkey" PRIMARY KEY, btree (id)
    "threads_local_repo_id_file_idx" btree (local_repo_id, file)

```

# Table "public.user_invite"
```
   Column   |           Type           |                        Modifiers                         
------------+--------------------------+----------------------------------------------------------
 uri        | text                     | 
 id         | integer                  | not null default nextval('user_invite_id_seq'::regclass)
 user_id    | text                     | 
 user_email | text                     | 
 org_id     | text                     | 
 org_name   | text                     | 
 sent_at    | timestamp with time zone | 
Indexes:
    "user_invite_pkey" PRIMARY KEY, btree (id)
    "user_invite_unique" UNIQUE, btree (user_id, org_id)

```
