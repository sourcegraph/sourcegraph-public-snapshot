# Table "public.comments"
```
     Column     |           Type           |                       Modifiers                       
----------------+--------------------------+-------------------------------------------------------
 id             | bigint                   | not null default nextval('comments_id_seq'::regclass)
 thread_id      | bigint                   | 
 contents       | text                     | 
 created_at     | timestamp with time zone | default now()
 updated_at     | timestamp with time zone | default now()
 deleted_at     | timestamp with time zone | 
 author_name    | text                     | 
 author_email   | text                     | 
 author_user_id | text                     | 
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

# Table "public.org_members"
```
    Column    |           Type           |                        Modifiers                         
--------------+--------------------------+----------------------------------------------------------
 id           | integer                  | not null default nextval('org_members_id_seq'::regclass)
 org_id       | integer                  | not null
 user_id      | text                     | not null
 email        | text                     | not null
 username     | text                     | not null
 created_at   | timestamp with time zone | default now()
 updated_at   | timestamp with time zone | default now()
 display_name | text                     | not null
 avatar_url   | text                     | 
Indexes:
    "org_members_pkey" PRIMARY KEY, btree (id)
    "org_members_org_id_user_email_key" UNIQUE CONSTRAINT, btree (org_id, email)
    "org_members_org_id_user_id_key" UNIQUE CONSTRAINT, btree (org_id, user_id)
    "org_members_org_id_user_name_key" UNIQUE CONSTRAINT, btree (org_id, username)
Foreign-key constraints:
    "org_members_references_orgs" FOREIGN KEY (org_id) REFERENCES orgs(id) ON DELETE RESTRICT

```

# Table "public.org_repos"
```
    Column    |           Type           |                        Modifiers                         
--------------+--------------------------+----------------------------------------------------------
 id           | bigint                   | not null default nextval('local_repos_id_seq'::regclass)
 remote_uri   | citext                   | 
 access_token | text                     | 
 created_at   | timestamp with time zone | default now()
 updated_at   | timestamp with time zone | default now()
 deleted_at   | timestamp with time zone | 
 org_id       | integer                  | 
Indexes:
    "local_repos_pkey" PRIMARY KEY, btree (id)
    "local_repos_remote_uri_idx" btree (remote_uri)

```

# Table "public.org_tags"
```
   Column   |           Type           |                       Modifiers                       
------------+--------------------------+-------------------------------------------------------
 id         | integer                  | not null default nextval('org_tags_id_seq'::regclass)
 org_id     | integer                  | not null
 name       | text                     | not null
 created_at | timestamp with time zone | default now()
 updated_at | timestamp with time zone | default now()
 deleted_at | timestamp with time zone | 
Indexes:
    "org_tags_pkey" PRIMARY KEY, btree (id)
Foreign-key constraints:
    "org_tags_references_users" FOREIGN KEY (org_id) REFERENCES orgs(id) ON DELETE RESTRICT

```

# Table "public.orgs"
```
    Column    |           Type           |                     Modifiers                     
--------------+--------------------------+---------------------------------------------------
 id           | integer                  | not null default nextval('orgs_id_seq'::regclass)
 name         | citext                   | not null
 created_at   | timestamp with time zone | default now()
 updated_at   | timestamp with time zone | default now()
 display_name | text                     | 
Indexes:
    "orgs_pkey" PRIMARY KEY, btree (id)
    "org_name_unique" UNIQUE CONSTRAINT, btree (name)
Check constraints:
    "org_display_name_valid" CHECK (char_length(display_name) <= 64)
    "org_name_valid_chars" CHECK (name ~ '^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,36}[a-zA-Z0-9])?$'::citext)
Referenced by:
    TABLE "org_members" CONSTRAINT "org_members_references_orgs" FOREIGN KEY (org_id) REFERENCES orgs(id) ON DELETE RESTRICT
    TABLE "org_tags" CONSTRAINT "org_tags_references_users" FOREIGN KEY (org_id) REFERENCES orgs(id) ON DELETE RESTRICT

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

# Table "public.schema_migrations"
```
 Column  |  Type   | Modifiers 
---------+---------+-----------
 version | bigint  | not null
 dirty   | boolean | not null
Indexes:
    "schema_migrations_pkey" PRIMARY KEY, btree (version)

```

# Table "public.threads"
```
     Column      |           Type           |                      Modifiers                       
-----------------+--------------------------+------------------------------------------------------
 id              | bigint                   | not null default nextval('threads_id_seq'::regclass)
 org_repo_id     | bigint                   | 
 file            | text                     | 
 revision        | text                     | 
 start_line      | integer                  | 
 end_line        | integer                  | 
 start_character | integer                  | 
 end_character   | integer                  | 
 created_at      | timestamp with time zone | default now()
 archived_at     | timestamp with time zone | 
 updated_at      | timestamp with time zone | default now()
 deleted_at      | timestamp with time zone | 
 range_length    | integer                  | 
 branch          | text                     | 
Indexes:
    "threads_pkey" PRIMARY KEY, btree (id)
    "threads_local_repo_id_file_idx" btree (org_repo_id, file)

```

# Table "public.user_tags"
```
   Column   |           Type           |                       Modifiers                        
------------+--------------------------+--------------------------------------------------------
 id         | integer                  | not null default nextval('user_tags_id_seq'::regclass)
 user_id    | integer                  | not null
 name       | text                     | not null
 created_at | timestamp with time zone | default now()
 updated_at | timestamp with time zone | default now()
 deleted_at | timestamp with time zone | 
Indexes:
    "user_tags_pkey" PRIMARY KEY, btree (id)
Foreign-key constraints:
    "user_tags_references_users" FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT

```

# Table "public.users"
```
    Column    |            Type             |                     Modifiers                      
--------------+-----------------------------+----------------------------------------------------
 id           | integer                     | not null default nextval('users_id_seq'::regclass)
 auth0_id     | text                        | not null
 email        | citext                      | not null
 username     | citext                      | not null
 display_name | text                        | not null
 avatar_url   | text                        | 
 created_at   | timestamp with time zone    | default now()
 updated_at   | timestamp with time zone    | default now()
 deleted_at   | timestamp without time zone | 
Indexes:
    "users_pkey" PRIMARY KEY, btree (id)
    "users_auth0_id_key" UNIQUE CONSTRAINT, btree (auth0_id)
    "users_email_key" UNIQUE CONSTRAINT, btree (email)
    "users_username_key" UNIQUE CONSTRAINT, btree (username)
Check constraints:
    "users_display_name_valid" CHECK (char_length(display_name) <= 64)
    "users_username_valid" CHECK (username ~ '^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,36}[a-zA-Z0-9])?$'::citext)
Referenced by:
    TABLE "user_tags" CONSTRAINT "user_tags_references_users" FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT

```
