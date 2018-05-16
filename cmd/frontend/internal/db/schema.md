# Table "public.access_tokens"
```
     Column      |           Type           |                         Modifiers                          
-----------------+--------------------------+------------------------------------------------------------
 id              | bigint                   | not null default nextval('access_tokens_id_seq'::regclass)
 subject_user_id | integer                  | not null
 value_sha256    | bytea                    | not null
 note            | text                     | not null
 created_at      | timestamp with time zone | not null default now()
 last_used_at    | timestamp with time zone | 
 deleted_at      | timestamp with time zone | 
 creator_user_id | integer                  | not null
 scopes          | text[]                   | not null
Indexes:
    "access_tokens_pkey" PRIMARY KEY, btree (id)
    "access_tokens_value_sha256_key" UNIQUE CONSTRAINT, btree (value_sha256)
    "access_tokens_lookup" hash (value_sha256) WHERE deleted_at IS NULL
Foreign-key constraints:
    "access_tokens_creator_user_id_fkey" FOREIGN KEY (creator_user_id) REFERENCES users(id)
    "access_tokens_subject_user_id_fkey" FOREIGN KEY (subject_user_id) REFERENCES users(id)

```

# Table "public.cert_cache"
```
   Column   |           Type           |                        Modifiers                        
------------+--------------------------+---------------------------------------------------------
 id         | bigint                   | not null default nextval('cert_cache_id_seq'::regclass)
 cache_key  | text                     | not null
 b64data    | text                     | not null
 created_at | timestamp with time zone | not null default now()
 updated_at | timestamp with time zone | not null default now()
 deleted_at | timestamp with time zone | 
Indexes:
    "cert_cache_pkey" PRIMARY KEY, btree (id)
    "cert_cache_key_idx" UNIQUE, btree (cache_key)

```

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
 author_user_id | integer                  | not null
Indexes:
    "comments_pkey" PRIMARY KEY, btree (id)
Foreign-key constraints:
    "comments_author_user_id_fkey" FOREIGN KEY (author_user_id) REFERENCES users(id) ON DELETE RESTRICT

```

# Table "public.comments_bkup_1514545501"
```
       Column       |           Type           | Modifiers 
--------------------+--------------------------+-----------
 id                 | bigint                   | 
 thread_id          | bigint                   | 
 contents           | text                     | 
 created_at         | timestamp with time zone | 
 updated_at         | timestamp with time zone | 
 deleted_at         | timestamp with time zone | 
 author_name        | text                     | 
 author_email       | text                     | 
 author_user_id_old | text                     | 
 author_user_id     | integer                  | 

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

# Table "public.org_members"
```
   Column   |           Type           |                        Modifiers                         
------------+--------------------------+----------------------------------------------------------
 id         | integer                  | not null default nextval('org_members_id_seq'::regclass)
 org_id     | integer                  | not null
 created_at | timestamp with time zone | not null default now()
 updated_at | timestamp with time zone | not null default now()
 user_id    | integer                  | not null
Indexes:
    "org_members_pkey" PRIMARY KEY, btree (id)
    "org_members_org_id_user_id_key" UNIQUE CONSTRAINT, btree (org_id, user_id)
Foreign-key constraints:
    "org_members_references_orgs" FOREIGN KEY (org_id) REFERENCES orgs(id) ON DELETE RESTRICT
    "org_members_user_id_fkey" FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT

```

# Table "public.org_members_bkup_1514536731"
```
   Column    |           Type           | Modifiers 
-------------+--------------------------+-----------
 id          | integer                  | 
 org_id      | integer                  | 
 user_id_old | text                     | 
 created_at  | timestamp with time zone | 
 updated_at  | timestamp with time zone | 
 user_id     | integer                  | 

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
 deleted_at        | timestamp with time zone | 
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
 description             | text                     | 
 language                | text                     | 
 fork                    | boolean                  | 
 created_at              | timestamp with time zone | not null default now()
 updated_at              | timestamp with time zone | 
 pushed_at               | timestamp with time zone | 
 indexed_revision        | text                     | 
 freeze_indexed_revision | boolean                  | 
 external_id             | text                     | 
 external_service_type   | text                     | 
 external_service_id     | text                     | 
 enabled                 | boolean                  | not null default true
Indexes:
    "repo_pkey" PRIMARY KEY, btree (id)
    "repo_uri_unique" UNIQUE, btree (uri)
    "repo_uri_trgm" gin (lower(uri::text) gin_trgm_ops)
Check constraints:
    "check_external" CHECK (external_id IS NULL AND external_service_type IS NULL AND external_service_id IS NULL OR external_id IS NOT NULL AND external_service_type IS NOT NULL AND external_service_id IS NOT NULL)
Referenced by:
    TABLE "global_dep" CONSTRAINT "global_dep_repo_id" FOREIGN KEY (repo_id) REFERENCES repo(id) ON DELETE RESTRICT
    TABLE "pkgs" CONSTRAINT "pkgs_repo_id" FOREIGN KEY (repo_id) REFERENCES repo(id) ON DELETE RESTRICT

```

# Table "public.saved_queries"
```
      Column      |           Type           | Modifiers 
------------------+--------------------------+-----------
 query            | text                     | not null
 last_executed    | timestamp with time zone | not null
 latest_result    | timestamp with time zone | not null
 exec_duration_ns | bigint                   | not null
Indexes:
    "saved_queries_query_unique" UNIQUE, btree (query)

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
 contents       | text                     | 
 created_at     | timestamp with time zone | not null default now()
 user_id        | integer                  | 
 author_user_id | integer                  | not null
Indexes:
    "settings_pkey" PRIMARY KEY, btree (id)
Foreign-key constraints:
    "settings_author_user_id_fkey" FOREIGN KEY (author_user_id) REFERENCES users(id) ON DELETE RESTRICT
    "settings_references_orgs" FOREIGN KEY (org_id) REFERENCES orgs(id) ON DELETE RESTRICT
    "settings_user_id_fkey" FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT

```

# Table "public.settings_bkup_1514702776"
```
       Column       |           Type           | Modifiers 
--------------------+--------------------------+-----------
 id                 | integer                  | 
 org_id             | integer                  | 
 author_user_id_old | text                     | 
 contents           | text                     | 
 created_at         | timestamp with time zone | 
 user_id            | integer                  | 
 author_user_id     | integer                  | 

```

# Table "public.shared_items"
```
     Column     |           Type           |                         Modifiers                         
----------------+--------------------------+-----------------------------------------------------------
 id             | bigint                   | not null default nextval('shared_items_id_seq'::regclass)
 ulid           | text                     | not null
 thread_id      | bigint                   | 
 comment_id     | bigint                   | 
 created_at     | timestamp with time zone | not null default now()
 updated_at     | timestamp with time zone | not null default now()
 deleted_at     | timestamp with time zone | 
 public         | boolean                  | not null default false
 author_user_id | integer                  | not null
Indexes:
    "shared_items_pkey" PRIMARY KEY, btree (id)
    "shared_items_ulid_idx" UNIQUE, btree (ulid)
Foreign-key constraints:
    "shared_items_author_user_id_fkey" FOREIGN KEY (author_user_id) REFERENCES users(id) ON DELETE RESTRICT

```

# Table "public.shared_items_bkup_1514546912"
```
       Column       |           Type           | Modifiers 
--------------------+--------------------------+-----------
 id                 | bigint                   | 
 ulid               | text                     | 
 author_user_id_old | text                     | 
 thread_id          | bigint                   | 
 comment_id         | bigint                   | 
 created_at         | timestamp with time zone | 
 updated_at         | timestamp with time zone | 
 deleted_at         | timestamp with time zone | 
 public             | boolean                  | 
 author_user_id     | integer                  | 

```

# Table "public.site_config"
```
   Column    |  Type   |       Modifiers        
-------------+---------+------------------------
 site_id     | uuid    | not null
 initialized | boolean | not null default false
Indexes:
    "site_config_pkey" PRIMARY KEY, btree (site_id)

```

# Table "public.survey_responses"
```
   Column   |           Type           |                           Modifiers                           
------------+--------------------------+---------------------------------------------------------------
 id         | bigint                   | not null default nextval('survey_responses_id_seq'::regclass)
 user_id    | integer                  | 
 email      | text                     | 
 score      | integer                  | not null
 reason     | text                     | 
 better     | text                     | 
 created_at | timestamp with time zone | not null default now()
Indexes:
    "survey_responses_pkey" PRIMARY KEY, btree (id)
Foreign-key constraints:
    "survey_responses_user_id_fkey" FOREIGN KEY (user_id) REFERENCES users(id)

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
 author_user_id                    | integer                  | not null
Indexes:
    "threads_pkey" PRIMARY KEY, btree (id)
    "threads_org_repo_id_branch_idx" btree (org_repo_id, branch)
    "threads_org_repo_id_lines_revision_path_branch_idx" btree (org_repo_id, lines_revision_path, branch)
    "threads_org_repo_id_repo_revision_path_branch_idx" btree (org_repo_id, repo_revision_path, branch)
Foreign-key constraints:
    "threads_author_user_id_fkey" FOREIGN KEY (author_user_id) REFERENCES users(id) ON DELETE RESTRICT

```

# Table "public.threads_bkup_1514544774"
```
              Column               |           Type           | Modifiers 
-----------------------------------+--------------------------+-----------
 id                                | bigint                   | 
 org_repo_id                       | bigint                   | 
 repo_revision_path                | text                     | 
 repo_revision                     | text                     | 
 start_line                        | integer                  | 
 end_line                          | integer                  | 
 start_character                   | integer                  | 
 end_character                     | integer                  | 
 created_at                        | timestamp with time zone | 
 archived_at                       | timestamp with time zone | 
 updated_at                        | timestamp with time zone | 
 deleted_at                        | timestamp with time zone | 
 range_length                      | integer                  | 
 branch                            | text                     | 
 author_user_id_old                | text                     | 
 html_lines_before                 | text                     | 
 html_lines                        | text                     | 
 html_lines_after                  | text                     | 
 text_lines_before                 | text                     | 
 text_lines                        | text                     | 
 text_lines_after                  | text                     | 
 text_lines_selection_range_start  | integer                  | 
 text_lines_selection_range_length | integer                  | 
 lines_revision                    | text                     | 
 lines_revision_path               | text                     | 
 author_user_id                    | integer                  | 

```

# Table "public.user_emails"
```
      Column       |           Type           |       Modifiers        
-------------------+--------------------------+------------------------
 user_id           | integer                  | not null
 email             | citext                   | not null
 created_at        | timestamp with time zone | not null default now()
 verification_code | text                     | 
 verified_at       | timestamp with time zone | 
Indexes:
    "user_emails_no_duplicates_per_user" UNIQUE CONSTRAINT, btree (user_id, email)
    "user_emails_unique_verified_email" EXCLUDE USING btree (email WITH =) WHERE (verified_at IS NOT NULL)
Foreign-key constraints:
    "user_emails_user_id_fkey" FOREIGN KEY (user_id) REFERENCES users(id)

```

# Table "public.user_external_accounts"
```
    Column    |           Type           |                              Modifiers                              
--------------+--------------------------+---------------------------------------------------------------------
 id           | integer                  | not null default nextval('user_external_accounts_id_seq'::regclass)
 user_id      | integer                  | not null
 service_type | text                     | not null
 service_id   | text                     | not null
 account_id   | text                     | not null
 auth_data    | jsonb                    | 
 account_data | jsonb                    | 
 created_at   | timestamp with time zone | not null default now()
 updated_at   | timestamp with time zone | not null default now()
 deleted_at   | timestamp with time zone | 
Indexes:
    "user_external_accounts_pkey" PRIMARY KEY, btree (id)
    "user_external_accounts_account" UNIQUE, btree (service_type, service_id, account_id) WHERE deleted_at IS NULL
Foreign-key constraints:
    "user_external_accounts_user_id_fkey" FOREIGN KEY (user_id) REFERENCES users(id)

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
 username          | citext                   | not null
 display_name      | text                     | 
 avatar_url        | text                     | 
 created_at        | timestamp with time zone | not null default now()
 updated_at        | timestamp with time zone | not null default now()
 deleted_at        | timestamp with time zone | 
 invite_quota      | integer                  | not null default 15
 passwd            | text                     | 
 passwd_reset_code | text                     | 
 passwd_reset_time | timestamp with time zone | 
 site_admin        | boolean                  | not null default false
 page_views        | integer                  | not null default 0
 search_queries    | integer                  | not null default 0
Indexes:
    "users_pkey" PRIMARY KEY, btree (id)
    "users_username" UNIQUE, btree (username) WHERE deleted_at IS NULL
Check constraints:
    "users_display_name_valid" CHECK (char_length(display_name) <= 64)
    "users_username_valid" CHECK (username ~ '^[a-zA-Z0-9](?:[a-zA-Z0-9]|-(?=[a-zA-Z0-9])){0,38}$'::citext)
Referenced by:
    TABLE "access_tokens" CONSTRAINT "access_tokens_creator_user_id_fkey" FOREIGN KEY (creator_user_id) REFERENCES users(id)
    TABLE "access_tokens" CONSTRAINT "access_tokens_subject_user_id_fkey" FOREIGN KEY (subject_user_id) REFERENCES users(id)
    TABLE "comments" CONSTRAINT "comments_author_user_id_fkey" FOREIGN KEY (author_user_id) REFERENCES users(id) ON DELETE RESTRICT
    TABLE "org_members" CONSTRAINT "org_members_user_id_fkey" FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT
    TABLE "settings" CONSTRAINT "settings_author_user_id_fkey" FOREIGN KEY (author_user_id) REFERENCES users(id) ON DELETE RESTRICT
    TABLE "settings" CONSTRAINT "settings_user_id_fkey" FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT
    TABLE "shared_items" CONSTRAINT "shared_items_author_user_id_fkey" FOREIGN KEY (author_user_id) REFERENCES users(id) ON DELETE RESTRICT
    TABLE "survey_responses" CONSTRAINT "survey_responses_user_id_fkey" FOREIGN KEY (user_id) REFERENCES users(id)
    TABLE "threads" CONSTRAINT "threads_author_user_id_fkey" FOREIGN KEY (author_user_id) REFERENCES users(id) ON DELETE RESTRICT
    TABLE "user_emails" CONSTRAINT "user_emails_user_id_fkey" FOREIGN KEY (user_id) REFERENCES users(id)
    TABLE "user_external_accounts" CONSTRAINT "user_external_accounts_user_id_fkey" FOREIGN KEY (user_id) REFERENCES users(id)
    TABLE "user_tags" CONSTRAINT "user_tags_references_users" FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT

```
