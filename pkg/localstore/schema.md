# Table "public.comments"
```
     Column     |           Type           |                       Modifiers                       
----------------+--------------------------+-------------------------------------------------------
 id             | bigint                   | not null default nextval('comments_id_seq'::regclass)
 thread_id      | bigint                   | 
 contents       | text                     | 
 created_at     | timestamp with time zone | not null default now()
 updated_at     | timestamp with time zone | not null default now()
 deleted_at     | timestamp with time zone | 
 author_name    | text                     | 
 author_email   | text                     | 
 author_user_id | text                     | 
Indexes:
    "comments_pkey" PRIMARY KEY, btree (id)

```

# Table "public.deployment_configuration"
```
      Column      |  Type   |  Modifiers   
------------------+---------+--------------
 id               | integer | not null
 app_id           | uuid    | not null
 enable_telemetry | boolean | default true
 email            | text    | 
 last_updated     | text    | 
Indexes:
    "deployment_configuration_pkey" PRIMARY KEY, btree (id)

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
Foreign-key constraints:
    "global_dep_repo_id" FOREIGN KEY (repo_id) REFERENCES repo(id) ON DELETE RESTRICT

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
   Column   |           Type           |                        Modifiers                         
------------+--------------------------+----------------------------------------------------------
 id         | integer                  | not null default nextval('org_members_id_seq'::regclass)
 org_id     | integer                  | not null
 user_id    | text                     | not null
 created_at | timestamp with time zone | not null default now()
 updated_at | timestamp with time zone | not null default now()
Indexes:
    "org_members_pkey" PRIMARY KEY, btree (id)
    "org_members_org_id_user_id_key" UNIQUE CONSTRAINT, btree (org_id, user_id)
Foreign-key constraints:
    "org_members_references_orgs" FOREIGN KEY (org_id) REFERENCES orgs(id) ON DELETE RESTRICT

```

# Table "public.org_repos"
```
       Column        |           Type           |                        Modifiers                         
---------------------+--------------------------+----------------------------------------------------------
 id                  | bigint                   | not null default nextval('local_repos_id_seq'::regclass)
 canonical_remote_id | citext                   | 
 created_at          | timestamp with time zone | not null default now()
 updated_at          | timestamp with time zone | not null default now()
 deleted_at          | timestamp with time zone | 
 org_id              | integer                  | 
 clone_url           | text                     | not null
Indexes:
    "local_repos_pkey" PRIMARY KEY, btree (id)
    "local_repos_remote_uri_idx" btree (canonical_remote_id)
Check constraints:
    "clone_url_valid" CHECK (clone_url ~ '^([^\s]+://)?[^\s]+$'::text)

```

# Table "public.org_tags"
```
   Column   |           Type           |                       Modifiers                       
------------+--------------------------+-------------------------------------------------------
 id         | integer                  | not null default nextval('org_tags_id_seq'::regclass)
 org_id     | integer                  | not null
 name       | text                     | not null
 created_at | timestamp with time zone | not null default now()
 updated_at | timestamp with time zone | not null default now()
 deleted_at | timestamp with time zone | 
Indexes:
    "org_tags_pkey" PRIMARY KEY, btree (id)
Foreign-key constraints:
    "org_tags_references_users" FOREIGN KEY (org_id) REFERENCES orgs(id) ON DELETE RESTRICT

```

# Table "public.orgs"
```
      Column       |           Type           |                     Modifiers                     
-------------------+--------------------------+---------------------------------------------------
 id                | integer                  | not null default nextval('orgs_id_seq'::regclass)
 name              | citext                   | not null
 created_at        | timestamp with time zone | not null default now()
 updated_at        | timestamp with time zone | not null default now()
 display_name      | text                     | 
 slack_webhook_url | text                     | 
Indexes:
    "orgs_pkey" PRIMARY KEY, btree (id)
    "org_name_unique" UNIQUE CONSTRAINT, btree (name)
Check constraints:
    "org_display_name_valid" CHECK (char_length(display_name) <= 64)
    "org_name_valid_chars" CHECK (name ~ '^[a-zA-Z0-9](?:[a-zA-Z0-9]|-(?=[a-zA-Z0-9])){0,38}$'::citext)
Referenced by:
    TABLE "org_members" CONSTRAINT "org_members_references_orgs" FOREIGN KEY (org_id) REFERENCES orgs(id) ON DELETE RESTRICT
    TABLE "org_tags" CONSTRAINT "org_tags_references_users" FOREIGN KEY (org_id) REFERENCES orgs(id) ON DELETE RESTRICT
    TABLE "settings" CONSTRAINT "settings_references_orgs" FOREIGN KEY (org_id) REFERENCES orgs(id) ON DELETE RESTRICT

```

# Table "public.phabricator_repos"
```
   Column   |           Type           |                           Modifiers                            
------------+--------------------------+----------------------------------------------------------------
 id         | integer                  | not null default nextval('phabricator_repos_id_seq'::regclass)
 callsign   | citext                   | not null
 uri        | citext                   | not null
 created_at | timestamp with time zone | not null default now()
 updated_at | timestamp with time zone | not null default now()
 deleted_at | timestamp with time zone | 
 url        | text                     | not null default ''::text
Indexes:
    "phabricator_repos_pkey" PRIMARY KEY, btree (id)
    "phabricator_repos_uri_key" UNIQUE CONSTRAINT, btree (uri)

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
Foreign-key constraints:
    "pkgs_repo_id" FOREIGN KEY (repo_id) REFERENCES repo(id) ON DELETE RESTRICT

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
 language                | text                     | 
 blocked                 | boolean                  | 
 fork                    | boolean                  | 
 private                 | boolean                  | 
 created_at              | timestamp with time zone | 
 updated_at              | timestamp with time zone | 
 pushed_at               | timestamp with time zone | 
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
Referenced by:
    TABLE "global_dep" CONSTRAINT "global_dep_repo_id" FOREIGN KEY (repo_id) REFERENCES repo(id) ON DELETE RESTRICT
    TABLE "pkgs" CONSTRAINT "pkgs_repo_id" FOREIGN KEY (repo_id) REFERENCES repo(id) ON DELETE RESTRICT

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

# Table "public.settings"
```
     Column     |           Type           |                       Modifiers                       
----------------+--------------------------+-------------------------------------------------------
 id             | integer                  | not null default nextval('settings_id_seq'::regclass)
 org_id         | integer                  | 
 author_auth_id | text                     | not null
 contents       | text                     | 
 created_at     | timestamp with time zone | not null default now()
 user_id        | integer                  | 
Indexes:
    "settings_pkey" PRIMARY KEY, btree (id)
Check constraints:
    "has_subject" CHECK (org_id IS NOT NULL OR user_id IS NOT NULL)
Foreign-key constraints:
    "settings_references_orgs" FOREIGN KEY (org_id) REFERENCES orgs(id) ON DELETE RESTRICT
    "settings_references_users" FOREIGN KEY (author_auth_id) REFERENCES users(auth_id) ON DELETE RESTRICT
    "settings_user_id_fkey" FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT

```

# Table "public.shared_items"
```
     Column     |           Type           |                         Modifiers                         
----------------+--------------------------+-----------------------------------------------------------
 id             | bigint                   | not null default nextval('shared_items_id_seq'::regclass)
 ulid           | text                     | not null
 author_user_id | text                     | not null
 thread_id      | bigint                   | 
 comment_id     | bigint                   | 
 created_at     | timestamp with time zone | not null default now()
 updated_at     | timestamp with time zone | not null default now()
 deleted_at     | timestamp with time zone | 
 public         | boolean                  | not null default false
Indexes:
    "shared_items_pkey" PRIMARY KEY, btree (id)
    "shared_items_ulid_idx" UNIQUE, btree (ulid)

```

# Table "public.threads"
```
              Column               |           Type           |                      Modifiers                       
-----------------------------------+--------------------------+------------------------------------------------------
 id                                | bigint                   | not null default nextval('threads_id_seq'::regclass)
 org_repo_id                       | bigint                   | 
 repo_revision_path                | text                     | not null
 repo_revision                     | text                     | not null
 start_line                        | integer                  | 
 end_line                          | integer                  | 
 start_character                   | integer                  | 
 end_character                     | integer                  | 
 created_at                        | timestamp with time zone | not null default now()
 archived_at                       | timestamp with time zone | 
 updated_at                        | timestamp with time zone | not null default now()
 deleted_at                        | timestamp with time zone | 
 range_length                      | integer                  | 
 branch                            | text                     | 
 author_user_id                    | text                     | not null
 html_lines_before                 | text                     | 
 html_lines                        | text                     | 
 html_lines_after                  | text                     | 
 text_lines_before                 | text                     | 
 text_lines                        | text                     | 
 text_lines_after                  | text                     | 
 text_lines_selection_range_start  | integer                  | not null default 0
 text_lines_selection_range_length | integer                  | not null default 0
 lines_revision                    | text                     | not null
 lines_revision_path               | text                     | not null
Indexes:
    "threads_pkey" PRIMARY KEY, btree (id)
    "threads_org_repo_id_branch_idx" btree (org_repo_id, branch)
    "threads_org_repo_id_lines_revision_path_branch_idx" btree (org_repo_id, lines_revision_path, branch)
    "threads_org_repo_id_repo_revision_path_branch_idx" btree (org_repo_id, repo_revision_path, branch)

```

# Table "public.user_activity"
```
     Column     |           Type           |                         Modifiers                          
----------------+--------------------------+------------------------------------------------------------
 id             | integer                  | not null default nextval('user_activity_id_seq'::regclass)
 user_id        | integer                  | not null
 page_views     | integer                  | not null default 0
 search_queries | integer                  | not null default 0
 created_at     | timestamp with time zone | not null default now()
 updated_at     | timestamp with time zone | not null default now()
Indexes:
    "user_activity_pkey" PRIMARY KEY, btree (id)
    "user_activity_user_id_key" UNIQUE CONSTRAINT, btree (user_id)
Foreign-key constraints:
    "user_activity" FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT

```

# Table "public.user_tags"
```
   Column   |           Type           |                       Modifiers                        
------------+--------------------------+--------------------------------------------------------
 id         | integer                  | not null default nextval('user_tags_id_seq'::regclass)
 user_id    | integer                  | not null
 name       | text                     | not null
 created_at | timestamp with time zone | not null default now()
 updated_at | timestamp with time zone | not null default now()
 deleted_at | timestamp with time zone | 
Indexes:
    "user_tags_pkey" PRIMARY KEY, btree (id)
Foreign-key constraints:
    "user_tags_references_users" FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT

```

# Table "public.users"
```
      Column       |           Type           |                     Modifiers                      
-------------------+--------------------------+----------------------------------------------------
 id                | integer                  | not null default nextval('users_id_seq'::regclass)
 auth_id           | text                     | not null
 email             | citext                   | not null
 username          | citext                   | not null
 display_name      | text                     | not null
 avatar_url        | text                     | 
 created_at        | timestamp with time zone | not null default now()
 updated_at        | timestamp with time zone | not null default now()
 deleted_at        | timestamp with time zone | 
 provider          | text                     | not null default ''::text
 invite_quota      | integer                  | not null default 15
 passwd            | text                     | 
 email_code        | text                     | 
 passwd_reset_code | text                     | 
 passwd_reset_time | timestamp with time zone | 
Indexes:
    "users_pkey" PRIMARY KEY, btree (id)
    "users_auth_id_key" UNIQUE CONSTRAINT, btree (auth_id)
    "users_email_key" UNIQUE CONSTRAINT, btree (email)
    "users_username_key" UNIQUE CONSTRAINT, btree (username)
Check constraints:
    "users_display_name_valid" CHECK (char_length(display_name) <= 64)
    "users_username_valid" CHECK (username ~ '^[a-zA-Z0-9](?:[a-zA-Z0-9]|-(?=[a-zA-Z0-9])){0,38}$'::citext)
Referenced by:
    TABLE "settings" CONSTRAINT "settings_references_users" FOREIGN KEY (author_auth_id) REFERENCES users(auth_id) ON DELETE RESTRICT
    TABLE "settings" CONSTRAINT "settings_user_id_fkey" FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT
    TABLE "user_activity" CONSTRAINT "user_activity" FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT
    TABLE "user_tags" CONSTRAINT "user_tags_references_users" FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT

```
