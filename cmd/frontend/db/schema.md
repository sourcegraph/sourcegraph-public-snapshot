# Table "public.access_tokens"
```
     Column      |           Type           | Collation | Nullable |                  Default                  
-----------------+--------------------------+-----------+----------+-------------------------------------------
 id              | bigint                   |           | not null | nextval('access_tokens_id_seq'::regclass)
 subject_user_id | integer                  |           | not null | 
 value_sha256    | bytea                    |           | not null | 
 note            | text                     |           | not null | 
 created_at      | timestamp with time zone |           | not null | now()
 last_used_at    | timestamp with time zone |           |          | 
 deleted_at      | timestamp with time zone |           |          | 
 creator_user_id | integer                  |           | not null | 
 scopes          | text[]                   |           | not null | 
Indexes:
    "access_tokens_pkey" PRIMARY KEY, btree (id)
    "access_tokens_value_sha256_key" UNIQUE CONSTRAINT, btree (value_sha256)
    "access_tokens_lookup" hash (value_sha256) WHERE deleted_at IS NULL
Foreign-key constraints:
    "access_tokens_creator_user_id_fkey" FOREIGN KEY (creator_user_id) REFERENCES users(id)
    "access_tokens_subject_user_id_fkey" FOREIGN KEY (subject_user_id) REFERENCES users(id)

```

# Table "public.comments"
```
     Column     |           Type           | Collation | Nullable |               Default                
----------------+--------------------------+-----------+----------+--------------------------------------
 id             | bigint                   |           | not null | nextval('comments_id_seq'::regclass)
 title          | text                     |           |          | 
 thread_id      | bigint                   |           | not null | 
 author_user_id | bigint                   |           |          | 
 body           | text                     |           | not null | 
 created_at     | timestamp with time zone |           | not null | now()
 updated_at     | timestamp with time zone |           | not null | now()
 deleted_at     | timestamp with time zone |           |          | 
Indexes:
    "comments_pkey" PRIMARY KEY, btree (id)
Foreign-key constraints:
    "comments_author_user_id_fkey" FOREIGN KEY (author_user_id) REFERENCES users(id) ON DELETE RESTRICT
    "comments_thread_id_fkey" FOREIGN KEY (thread_id) REFERENCES threads(id) ON DELETE RESTRICT

```

# Table "public.critical_and_site_config"
```
   Column   |           Type           | Collation | Nullable |                       Default                        
------------+--------------------------+-----------+----------+------------------------------------------------------
 id         | integer                  |           | not null | nextval('critical_and_site_config_id_seq'::regclass)
 type       | critical_or_site         |           | not null | 
 contents   | text                     |           | not null | 
 created_at | timestamp with time zone |           | not null | now()
 updated_at | timestamp with time zone |           | not null | now()
Indexes:
    "critical_and_site_config_pkey" PRIMARY KEY, btree (id)
    "critical_and_site_config_unique" UNIQUE, btree (id, type)

```

# Table "public.discussion_comments"
```
     Column     |           Type           | Collation | Nullable |                     Default                     
----------------+--------------------------+-----------+----------+-------------------------------------------------
 id             | bigint                   |           | not null | nextval('discussion_comments_id_seq'::regclass)
 thread_id      | bigint                   |           | not null | 
 author_user_id | integer                  |           | not null | 
 contents       | text                     |           | not null | 
 created_at     | timestamp with time zone |           | not null | now()
 updated_at     | timestamp with time zone |           | not null | now()
 deleted_at     | timestamp with time zone |           |          | 
 reports        | text[]                   |           | not null | '{}'::text[]
Indexes:
    "discussion_comments_pkey" PRIMARY KEY, btree (id)
    "discussion_comments_author_user_id_idx" btree (author_user_id)
    "discussion_comments_reports_array_length_idx" btree (array_length(reports, 1))
    "discussion_comments_thread_id_idx" btree (thread_id)
Foreign-key constraints:
    "discussion_comments_author_user_id_fkey" FOREIGN KEY (author_user_id) REFERENCES users(id) ON DELETE RESTRICT
    "discussion_comments_thread_id_fkey" FOREIGN KEY (thread_id) REFERENCES discussion_threads(id) ON DELETE CASCADE

```

# Table "public.discussion_mail_reply_tokens"
```
   Column   |           Type           | Collation | Nullable | Default 
------------+--------------------------+-----------+----------+---------
 token      | text                     |           | not null | 
 user_id    | integer                  |           | not null | 
 thread_id  | bigint                   |           | not null | 
 deleted_at | timestamp with time zone |           |          | 
Indexes:
    "discussion_mail_reply_tokens_pkey" PRIMARY KEY, btree (token)
    "discussion_mail_reply_tokens_token_idx" btree (token)
    "discussion_mail_reply_tokens_user_id_thread_id_idx" btree (user_id, thread_id)
Foreign-key constraints:
    "discussion_mail_reply_tokens_thread_id_fkey" FOREIGN KEY (thread_id) REFERENCES discussion_threads(id) ON DELETE CASCADE
    "discussion_mail_reply_tokens_user_id_fkey" FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT

```

# Table "public.discussion_threads"
```
     Column     |           Type           | Collation | Nullable |                    Default                     
----------------+--------------------------+-----------+----------+------------------------------------------------
 id             | bigint                   |           | not null | nextval('discussion_threads_id_seq'::regclass)
 author_user_id | integer                  |           | not null | 
 title          | text                     |           |          | 
 target_repo_id | bigint                   |           |          | 
 created_at     | timestamp with time zone |           | not null | now()
 archived_at    | timestamp with time zone |           |          | 
 updated_at     | timestamp with time zone |           | not null | now()
 deleted_at     | timestamp with time zone |           |          | 
Indexes:
    "discussion_threads_pkey" PRIMARY KEY, btree (id)
    "discussion_threads_author_user_id_idx" btree (author_user_id)
    "discussion_threads_id_idx" btree (id)
Foreign-key constraints:
    "discussion_threads_author_user_id_fkey" FOREIGN KEY (author_user_id) REFERENCES users(id) ON DELETE RESTRICT
    "discussion_threads_target_repo_id_fk" FOREIGN KEY (target_repo_id) REFERENCES discussion_threads_target_repo(id) ON DELETE CASCADE
Referenced by:
    TABLE "discussion_comments" CONSTRAINT "discussion_comments_thread_id_fkey" FOREIGN KEY (thread_id) REFERENCES discussion_threads(id) ON DELETE CASCADE
    TABLE "discussion_mail_reply_tokens" CONSTRAINT "discussion_mail_reply_tokens_thread_id_fkey" FOREIGN KEY (thread_id) REFERENCES discussion_threads(id) ON DELETE CASCADE
    TABLE "discussion_threads_target_repo" CONSTRAINT "discussion_threads_target_repo_thread_id_fkey" FOREIGN KEY (thread_id) REFERENCES discussion_threads(id) ON DELETE CASCADE

```

# Table "public.discussion_threads_target_repo"
```
     Column      |  Type   | Collation | Nullable |                          Default                           
-----------------+---------+-----------+----------+------------------------------------------------------------
 id              | bigint  |           | not null | nextval('discussion_threads_target_repo_id_seq'::regclass)
 thread_id       | bigint  |           | not null | 
 repo_id         | integer |           | not null | 
 path            | text    |           |          | 
 branch          | text    |           |          | 
 revision        | text    |           |          | 
 start_line      | integer |           |          | 
 end_line        | integer |           |          | 
 start_character | integer |           |          | 
 end_character   | integer |           |          | 
 lines_before    | text    |           |          | 
 lines           | text    |           |          | 
 lines_after     | text    |           |          | 
Indexes:
    "discussion_threads_target_repo_pkey" PRIMARY KEY, btree (id)
    "discussion_threads_target_repo_repo_id_path_idx" btree (repo_id, path)
Foreign-key constraints:
    "discussion_threads_target_repo_repo_id_fkey" FOREIGN KEY (repo_id) REFERENCES repo(id) ON DELETE CASCADE
    "discussion_threads_target_repo_thread_id_fkey" FOREIGN KEY (thread_id) REFERENCES discussion_threads(id) ON DELETE CASCADE
Referenced by:
    TABLE "discussion_threads" CONSTRAINT "discussion_threads_target_repo_id_fk" FOREIGN KEY (target_repo_id) REFERENCES discussion_threads_target_repo(id) ON DELETE CASCADE

```

# Table "public.external_services"
```
    Column    |           Type           | Collation | Nullable |                    Default                    
--------------+--------------------------+-----------+----------+-----------------------------------------------
 id           | bigint                   |           | not null | nextval('external_services_id_seq'::regclass)
 kind         | text                     |           | not null | 
 display_name | text                     |           | not null | 
 config       | text                     |           | not null | 
 created_at   | timestamp with time zone |           | not null | now()
 updated_at   | timestamp with time zone |           | not null | now()
 deleted_at   | timestamp with time zone |           |          | 
Indexes:
    "external_services_pkey" PRIMARY KEY, btree (id)
Check constraints:
    "check_non_empty_config" CHECK (btrim(config) <> ''::text)

```

# Table "public.global_state"
```
         Column          |  Type   | Collation | Nullable | Default  
-------------------------+---------+-----------+----------+----------
 site_id                 | uuid    |           | not null | 
 initialized             | boolean |           | not null | false
 mgmt_password_plaintext | text    |           | not null | ''::text
 mgmt_password_bcrypt    | text    |           | not null | ''::text
Indexes:
    "global_state_pkey" PRIMARY KEY, btree (site_id)

```

# Table "public.labels"
```
   Column    |  Type  | Collation | Nullable |              Default               
-------------+--------+-----------+----------+------------------------------------
 id          | bigint |           | not null | nextval('labels_id_seq'::regclass)
 name        | citext |           | not null | 
 description | text   |           |          | 
 color_hex   | text   |           | not null | 
Indexes:
    "labels_pkey" PRIMARY KEY, btree (id)
Referenced by:
    TABLE "threads_labels" CONSTRAINT "threads_labels_label_id_fkey" FOREIGN KEY (label_id) REFERENCES labels(id) ON DELETE CASCADE

```

# Table "public.names"
```
 Column  |  Type   | Collation | Nullable | Default 
---------+---------+-----------+----------+---------
 name    | citext  |           | not null | 
 user_id | integer |           |          | 
 org_id  | integer |           |          | 
Indexes:
    "names_pkey" PRIMARY KEY, btree (name)
Check constraints:
    "names_check" CHECK (user_id IS NOT NULL OR org_id IS NOT NULL)
Foreign-key constraints:
    "names_org_id_fkey" FOREIGN KEY (org_id) REFERENCES orgs(id) ON UPDATE CASCADE ON DELETE CASCADE
    "names_user_id_fkey" FOREIGN KEY (user_id) REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE

```

# Table "public.org_invitations"
```
      Column       |           Type           | Collation | Nullable |                   Default                   
-------------------+--------------------------+-----------+----------+---------------------------------------------
 id                | bigint                   |           | not null | nextval('org_invitations_id_seq'::regclass)
 org_id            | integer                  |           | not null | 
 sender_user_id    | integer                  |           | not null | 
 recipient_user_id | integer                  |           | not null | 
 created_at        | timestamp with time zone |           | not null | now()
 notified_at       | timestamp with time zone |           |          | 
 responded_at      | timestamp with time zone |           |          | 
 response_type     | boolean                  |           |          | 
 revoked_at        | timestamp with time zone |           |          | 
 deleted_at        | timestamp with time zone |           |          | 
Indexes:
    "org_invitations_pkey" PRIMARY KEY, btree (id)
    "org_invitations_singleflight" UNIQUE, btree (org_id, recipient_user_id) WHERE responded_at IS NULL AND revoked_at IS NULL AND deleted_at IS NULL
    "org_invitations_org_id" btree (org_id) WHERE deleted_at IS NULL
    "org_invitations_recipient_user_id" btree (recipient_user_id) WHERE deleted_at IS NULL
Check constraints:
    "check_atomic_response" CHECK ((responded_at IS NULL) = (response_type IS NULL))
    "check_single_use" CHECK (responded_at IS NULL AND response_type IS NULL OR revoked_at IS NULL)
Foreign-key constraints:
    "org_invitations_org_id_fkey" FOREIGN KEY (org_id) REFERENCES orgs(id)
    "org_invitations_recipient_user_id_fkey" FOREIGN KEY (recipient_user_id) REFERENCES users(id)
    "org_invitations_sender_user_id_fkey" FOREIGN KEY (sender_user_id) REFERENCES users(id)

```

# Table "public.org_members"
```
   Column   |           Type           | Collation | Nullable |                 Default                 
------------+--------------------------+-----------+----------+-----------------------------------------
 id         | integer                  |           | not null | nextval('org_members_id_seq'::regclass)
 org_id     | integer                  |           | not null | 
 created_at | timestamp with time zone |           | not null | now()
 updated_at | timestamp with time zone |           | not null | now()
 user_id    | integer                  |           | not null | 
Indexes:
    "org_members_pkey" PRIMARY KEY, btree (id)
    "org_members_org_id_user_id_key" UNIQUE CONSTRAINT, btree (org_id, user_id)
Foreign-key constraints:
    "org_members_references_orgs" FOREIGN KEY (org_id) REFERENCES orgs(id) ON DELETE RESTRICT
    "org_members_user_id_fkey" FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT

```

# Table "public.org_members_bkup_1514536731"
```
   Column    |           Type           | Collation | Nullable | Default 
-------------+--------------------------+-----------+----------+---------
 id          | integer                  |           |          | 
 org_id      | integer                  |           |          | 
 user_id_old | text                     |           |          | 
 created_at  | timestamp with time zone |           |          | 
 updated_at  | timestamp with time zone |           |          | 
 user_id     | integer                  |           |          | 

```

# Table "public.orgs"
```
      Column       |           Type           | Collation | Nullable |             Default              
-------------------+--------------------------+-----------+----------+----------------------------------
 id                | integer                  |           | not null | nextval('orgs_id_seq'::regclass)
 name              | citext                   |           | not null | 
 created_at        | timestamp with time zone |           | not null | now()
 updated_at        | timestamp with time zone |           | not null | now()
 display_name      | text                     |           |          | 
 slack_webhook_url | text                     |           |          | 
 deleted_at        | timestamp with time zone |           |          | 
Indexes:
    "orgs_pkey" PRIMARY KEY, btree (id)
    "orgs_name" UNIQUE, btree (name) WHERE deleted_at IS NULL
Check constraints:
    "orgs_display_name_max_length" CHECK (char_length(display_name) <= 255)
    "orgs_name_max_length" CHECK (char_length(name::text) <= 255)
    "orgs_name_valid_chars" CHECK (name ~ '^[a-zA-Z0-9](?:[a-zA-Z0-9]|-(?=[a-zA-Z0-9]))*$'::citext)
Referenced by:
    TABLE "names" CONSTRAINT "names_org_id_fkey" FOREIGN KEY (org_id) REFERENCES orgs(id) ON UPDATE CASCADE ON DELETE CASCADE
    TABLE "org_invitations" CONSTRAINT "org_invitations_org_id_fkey" FOREIGN KEY (org_id) REFERENCES orgs(id)
    TABLE "org_members" CONSTRAINT "org_members_references_orgs" FOREIGN KEY (org_id) REFERENCES orgs(id) ON DELETE RESTRICT
    TABLE "registry_extensions" CONSTRAINT "registry_extensions_publisher_org_id_fkey" FOREIGN KEY (publisher_org_id) REFERENCES orgs(id)
    TABLE "saved_searches" CONSTRAINT "saved_searches_org_id_fkey" FOREIGN KEY (org_id) REFERENCES orgs(id)
    TABLE "settings" CONSTRAINT "settings_references_orgs" FOREIGN KEY (org_id) REFERENCES orgs(id) ON DELETE RESTRICT

```

# Table "public.phabricator_repos"
```
   Column   |           Type           | Collation | Nullable |                    Default                    
------------+--------------------------+-----------+----------+-----------------------------------------------
 id         | integer                  |           | not null | nextval('phabricator_repos_id_seq'::regclass)
 callsign   | citext                   |           | not null | 
 repo_name  | citext                   |           | not null | 
 created_at | timestamp with time zone |           | not null | now()
 updated_at | timestamp with time zone |           | not null | now()
 deleted_at | timestamp with time zone |           |          | 
 url        | text                     |           | not null | ''::text
Indexes:
    "phabricator_repos_pkey" PRIMARY KEY, btree (id)
    "phabricator_repos_repo_name_key" UNIQUE CONSTRAINT, btree (repo_name)

```

# Table "public.product_licenses"
```
         Column          |           Type           | Collation | Nullable | Default 
-------------------------+--------------------------+-----------+----------+---------
 id                      | uuid                     |           | not null | 
 product_subscription_id | uuid                     |           | not null | 
 license_key             | text                     |           | not null | 
 created_at              | timestamp with time zone |           | not null | now()
Indexes:
    "product_licenses_pkey" PRIMARY KEY, btree (id)
Foreign-key constraints:
    "product_licenses_product_subscription_id_fkey" FOREIGN KEY (product_subscription_id) REFERENCES product_subscriptions(id)

```

# Table "public.product_subscriptions"
```
         Column          |           Type           | Collation | Nullable | Default 
-------------------------+--------------------------+-----------+----------+---------
 id                      | uuid                     |           | not null | 
 user_id                 | integer                  |           | not null | 
 billing_subscription_id | text                     |           |          | 
 created_at              | timestamp with time zone |           | not null | now()
 updated_at              | timestamp with time zone |           | not null | now()
 archived_at             | timestamp with time zone |           |          | 
Indexes:
    "product_subscriptions_pkey" PRIMARY KEY, btree (id)
Foreign-key constraints:
    "product_subscriptions_user_id_fkey" FOREIGN KEY (user_id) REFERENCES users(id)
Referenced by:
    TABLE "product_licenses" CONSTRAINT "product_licenses_product_subscription_id_fkey" FOREIGN KEY (product_subscription_id) REFERENCES product_subscriptions(id)

```

# Table "public.query_runner_state"
```
      Column      |           Type           | Modifiers 
------------------+--------------------------+-----------
 query            | text                     | 
 last_executed    | timestamp with time zone | 
 latest_result    | timestamp with time zone | 
 exec_duration_ns | bigint                   | 

```

# Table "public.recent_searches"
```
   Column   |            Type             | Collation | Nullable |                   Default                   
------------+-----------------------------+-----------+----------+---------------------------------------------
 id         | integer                     |           | not null | nextval('recent_searches_id_seq'::regclass)
 query      | text                        |           | not null | 
 created_at | timestamp without time zone |           | not null | now()
Indexes:
    "recent_searches_pkey" PRIMARY KEY, btree (id)

```

# Table "public.registry_extension_releases"
```
        Column         |           Type           | Collation | Nullable |                         Default                         
-----------------------+--------------------------+-----------+----------+---------------------------------------------------------
 id                    | bigint                   |           | not null | nextval('registry_extension_releases_id_seq'::regclass)
 registry_extension_id | integer                  |           | not null | 
 creator_user_id       | integer                  |           | not null | 
 release_version       | citext                   |           |          | 
 release_tag           | citext                   |           | not null | 
 manifest              | jsonb                    |           | not null | 
 bundle                | text                     |           |          | 
 created_at            | timestamp with time zone |           | not null | now()
 deleted_at            | timestamp with time zone |           |          | 
 source_map            | text                     |           |          | 
Indexes:
    "registry_extension_releases_pkey" PRIMARY KEY, btree (id)
    "registry_extension_releases_version" UNIQUE, btree (registry_extension_id, release_version) WHERE release_version IS NOT NULL
    "registry_extension_releases_registry_extension_id" btree (registry_extension_id, release_tag, created_at DESC) WHERE deleted_at IS NULL
Foreign-key constraints:
    "registry_extension_releases_creator_user_id_fkey" FOREIGN KEY (creator_user_id) REFERENCES users(id)
    "registry_extension_releases_registry_extension_id_fkey" FOREIGN KEY (registry_extension_id) REFERENCES registry_extensions(id) ON UPDATE CASCADE ON DELETE CASCADE

```

# Table "public.registry_extensions"
```
      Column       |           Type           | Collation | Nullable |                     Default                     
-------------------+--------------------------+-----------+----------+-------------------------------------------------
 id                | integer                  |           | not null | nextval('registry_extensions_id_seq'::regclass)
 uuid              | uuid                     |           | not null | 
 publisher_user_id | integer                  |           |          | 
 publisher_org_id  | integer                  |           |          | 
 name              | citext                   |           | not null | 
 manifest          | text                     |           |          | 
 created_at        | timestamp with time zone |           | not null | now()
 updated_at        | timestamp with time zone |           | not null | now()
 deleted_at        | timestamp with time zone |           |          | 
Indexes:
    "registry_extensions_pkey" PRIMARY KEY, btree (id)
    "registry_extensions_publisher_name" UNIQUE, btree (COALESCE(publisher_user_id, 0), COALESCE(publisher_org_id, 0), name) WHERE deleted_at IS NULL
    "registry_extensions_uuid" UNIQUE, btree (uuid)
Check constraints:
    "registry_extensions_name_length" CHECK (char_length(name::text) > 0 AND char_length(name::text) <= 128)
    "registry_extensions_name_valid_chars" CHECK (name ~ '^[a-zA-Z0-9](?:[a-zA-Z0-9]|[_.-](?=[a-zA-Z0-9]))*$'::citext)
    "registry_extensions_single_publisher" CHECK ((publisher_user_id IS NULL) <> (publisher_org_id IS NULL))
Foreign-key constraints:
    "registry_extensions_publisher_org_id_fkey" FOREIGN KEY (publisher_org_id) REFERENCES orgs(id)
    "registry_extensions_publisher_user_id_fkey" FOREIGN KEY (publisher_user_id) REFERENCES users(id)
Referenced by:
    TABLE "registry_extension_releases" CONSTRAINT "registry_extension_releases_registry_extension_id_fkey" FOREIGN KEY (registry_extension_id) REFERENCES registry_extensions(id) ON UPDATE CASCADE ON DELETE CASCADE

```

# Table "public.repo"
```
        Column         |           Type           | Collation | Nullable |             Default              
-----------------------+--------------------------+-----------+----------+----------------------------------
 id                    | integer                  |           | not null | nextval('repo_id_seq'::regclass)
 name                  | citext                   |           | not null | 
 description           | text                     |           |          | 
 language              | text                     |           |          | 
 fork                  | boolean                  |           |          | 
 created_at            | timestamp with time zone |           | not null | now()
 updated_at            | timestamp with time zone |           |          | 
 external_id           | text                     |           |          | 
 external_service_type | text                     |           |          | 
 external_service_id   | text                     |           |          | 
 enabled               | boolean                  |           | not null | true
 archived              | boolean                  |           | not null | false
 uri                   | citext                   |           |          | 
 deleted_at            | timestamp with time zone |           |          | 
 sources               | jsonb                    |           | not null | '{}'::jsonb
 metadata              | jsonb                    |           | not null | '{}'::jsonb
Indexes:
    "repo_pkey" PRIMARY KEY, btree (id)
    "repo_external_service_unique_idx" UNIQUE, btree (external_service_type, external_service_id, external_id) WHERE external_service_type IS NOT NULL AND external_service_id IS NOT NULL AND external_id IS NOT NULL
    "repo_name_unique" UNIQUE CONSTRAINT, btree (name) DEFERRABLE
    "repo_metadata_gin_idx" gin (metadata)
    "repo_name_trgm" gin (lower(name::text) gin_trgm_ops)
    "repo_sources_gin_idx" gin (sources)
    "repo_uri_idx" btree (uri)
Check constraints:
    "check_name_nonempty" CHECK (name <> ''::citext)
    "deleted_at_unused" CHECK (deleted_at IS NULL)
    "repo_metadata_check" CHECK (jsonb_typeof(metadata) = 'object'::text)
    "repo_sources_check" CHECK (jsonb_typeof(sources) = 'object'::text)
Referenced by:
    TABLE "discussion_threads_target_repo" CONSTRAINT "discussion_threads_target_repo_repo_id_fkey" FOREIGN KEY (repo_id) REFERENCES repo(id) ON DELETE CASCADE

```

# Table "public.saved_queries"
```
      Column      |           Type           | Collation | Nullable | Default 
------------------+--------------------------+-----------+----------+---------
 query            | text                     |           | not null | 
 last_executed    | timestamp with time zone |           | not null | 
 latest_result    | timestamp with time zone |           | not null | 
 exec_duration_ns | bigint                   |           | not null | 
Indexes:
    "saved_queries_query_unique" UNIQUE, btree (query)

```

# Table "public.saved_searches"
```
      Column       |           Type           |                          Modifiers                          
-------------------+--------------------------+-------------------------------------------------------------
 id                | integer                  | not null default nextval('saved_searches_id_seq'::regclass)
 description       | text                     | not null
 query             | text                     | not null
 created_at        | timestamp with time zone | not null default now()
 updated_at        | timestamp with time zone | not null default now()
 notify_owner      | boolean                  | not null
 notify_slack      | boolean                  | not null
 user_id           | integer                  | 
 org_id            | integer                  | 
 slack_webhook_url | text                     | 
Indexes:
    "saved_searches_pkey" PRIMARY KEY, btree (id)
Check constraints:
    "user_or_org_id_not_null" CHECK (user_id IS NOT NULL AND org_id IS NULL OR org_id IS NOT NULL AND user_id IS NULL)
Foreign-key constraints:
    "saved_searches_org_id_fkey" FOREIGN KEY (org_id) REFERENCES orgs(id)
    "saved_searches_user_id_fkey" FOREIGN KEY (user_id) REFERENCES users(id)

```

# Table "public.schema_migrations"
```
 Column  |  Type   | Collation | Nullable | Default 
---------+---------+-----------+----------+---------
 version | bigint  |           | not null | 
 dirty   | boolean |           | not null | 
Indexes:
    "schema_migrations_pkey" PRIMARY KEY, btree (version)

```

# Table "public.settings"
```
     Column     |           Type           | Collation | Nullable |               Default                
----------------+--------------------------+-----------+----------+--------------------------------------
 id             | integer                  |           | not null | nextval('settings_id_seq'::regclass)
 org_id         | integer                  |           |          | 
 contents       | text                     |           |          | 
 created_at     | timestamp with time zone |           | not null | now()
 user_id        | integer                  |           |          | 
 author_user_id | integer                  |           |          | 
Indexes:
    "settings_pkey" PRIMARY KEY, btree (id)
Foreign-key constraints:
    "settings_author_user_id_fkey" FOREIGN KEY (author_user_id) REFERENCES users(id) ON DELETE RESTRICT
    "settings_references_orgs" FOREIGN KEY (org_id) REFERENCES orgs(id) ON DELETE RESTRICT
    "settings_user_id_fkey" FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT

```

# Table "public.settings_bkup_1514702776"
```
       Column       |           Type           | Collation | Nullable | Default 
--------------------+--------------------------+-----------+----------+---------
 id                 | integer                  |           |          | 
 org_id             | integer                  |           |          | 
 author_user_id_old | text                     |           |          | 
 contents           | text                     |           |          | 
 created_at         | timestamp with time zone |           |          | 
 user_id            | integer                  |           |          | 
 author_user_id     | integer                  |           |          | 

```

# Table "public.survey_responses"
```
   Column   |           Type           | Collation | Nullable |                   Default                    
------------+--------------------------+-----------+----------+----------------------------------------------
 id         | bigint                   |           | not null | nextval('survey_responses_id_seq'::regclass)
 user_id    | integer                  |           |          | 
 email      | text                     |           |          | 
 score      | integer                  |           | not null | 
 reason     | text                     |           |          | 
 better     | text                     |           |          | 
 created_at | timestamp with time zone |           | not null | now()
Indexes:
    "survey_responses_pkey" PRIMARY KEY, btree (id)
Foreign-key constraints:
    "survey_responses_user_id_fkey" FOREIGN KEY (user_id) REFERENCES users(id)

```

# Table "public.threads"
```
      Column      |           Type           | Collation | Nullable |               Default               
------------------+--------------------------+-----------+----------+-------------------------------------
 id               | bigint                   |           | not null | nextval('threads_id_seq'::regclass)
 kind             | text                     |           | not null | 
 settings         | text                     |           |          | 
 creator_user_id  | bigint                   |           | not null | 
 parent_thread_id | bigint                   |           |          | 
 is_open          | boolean                  |           | not null | true
 created_at       | timestamp with time zone |           | not null | now()
 updated_at       | timestamp with time zone |           | not null | now()
 deleted_at       | timestamp with time zone |           |          | 
Indexes:
    "threads_pkey" PRIMARY KEY, btree (id)
Foreign-key constraints:
    "threads_creator_user_id_fkey" FOREIGN KEY (creator_user_id) REFERENCES users(id) ON DELETE RESTRICT
    "threads_parent_thread_id_fkey" FOREIGN KEY (parent_thread_id) REFERENCES threads(id) ON DELETE RESTRICT
Referenced by:
    TABLE "comments" CONSTRAINT "comments_thread_id_fkey" FOREIGN KEY (thread_id) REFERENCES threads(id) ON DELETE RESTRICT
    TABLE "threads_assignees" CONSTRAINT "threads_assignees_thread_id_fkey" FOREIGN KEY (thread_id) REFERENCES threads(id) ON DELETE CASCADE
    TABLE "threads_labels" CONSTRAINT "threads_labels_thread_id_fkey" FOREIGN KEY (thread_id) REFERENCES threads(id) ON DELETE CASCADE
    TABLE "threads" CONSTRAINT "threads_parent_thread_id_fkey" FOREIGN KEY (parent_thread_id) REFERENCES threads(id) ON DELETE RESTRICT

```

# Table "public.threads_assignees"
```
  Column   |  Type  | Collation | Nullable | Default 
-----------+--------+-----------+----------+---------
 thread_id | bigint |           | not null | 
 user_id   | bigint |           | not null | 
Foreign-key constraints:
    "threads_assignees_thread_id_fkey" FOREIGN KEY (thread_id) REFERENCES threads(id) ON DELETE CASCADE
    "threads_assignees_user_id_fkey" FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE

```

# Table "public.threads_labels"
```
  Column   |  Type  | Collation | Nullable | Default 
-----------+--------+-----------+----------+---------
 thread_id | bigint |           | not null | 
 label_id  | bigint |           | not null | 
Foreign-key constraints:
    "threads_labels_label_id_fkey" FOREIGN KEY (label_id) REFERENCES labels(id) ON DELETE CASCADE
    "threads_labels_thread_id_fkey" FOREIGN KEY (thread_id) REFERENCES threads(id) ON DELETE CASCADE

```

# Table "public.user_emails"
```
      Column       |           Type           | Collation | Nullable | Default 
-------------------+--------------------------+-----------+----------+---------
 user_id           | integer                  |           | not null | 
 email             | citext                   |           | not null | 
 created_at        | timestamp with time zone |           | not null | now()
 verification_code | text                     |           |          | 
 verified_at       | timestamp with time zone |           |          | 
Indexes:
    "user_emails_no_duplicates_per_user" UNIQUE CONSTRAINT, btree (user_id, email)
    "user_emails_unique_verified_email" EXCLUDE USING btree (email WITH =) WHERE (verified_at IS NOT NULL)
Foreign-key constraints:
    "user_emails_user_id_fkey" FOREIGN KEY (user_id) REFERENCES users(id)

```

# Table "public.user_external_accounts"
```
    Column    |           Type           | Collation | Nullable |                      Default                       
--------------+--------------------------+-----------+----------+----------------------------------------------------
 id           | integer                  |           | not null | nextval('user_external_accounts_id_seq'::regclass)
 user_id      | integer                  |           | not null | 
 service_type | text                     |           | not null | 
 service_id   | text                     |           | not null | 
 account_id   | text                     |           | not null | 
 auth_data    | jsonb                    |           |          | 
 account_data | jsonb                    |           |          | 
 created_at   | timestamp with time zone |           | not null | now()
 updated_at   | timestamp with time zone |           | not null | now()
 deleted_at   | timestamp with time zone |           |          | 
 client_id    | text                     |           | not null | 
Indexes:
    "user_external_accounts_pkey" PRIMARY KEY, btree (id)
    "user_external_accounts_account" UNIQUE, btree (service_type, service_id, client_id, account_id) WHERE deleted_at IS NULL
Foreign-key constraints:
    "user_external_accounts_user_id_fkey" FOREIGN KEY (user_id) REFERENCES users(id)

```

# Table "public.user_permissions"
```
   Column    |           Type           | Modifiers 
-------------+--------------------------+-----------
 user_id     | integer                  | not null
 permission  | text                     | not null
 object_type | text                     | not null
 object_ids  | bytea                    | not null
 updated_at  | timestamp with time zone | not null
Indexes:
    "user_permissions_perm_object_unique" UNIQUE CONSTRAINT, btree (user_id, permission, object_type)

```

# Table "public.users"
```
       Column        |           Type           | Collation | Nullable |              Default              
---------------------+--------------------------+-----------+----------+-----------------------------------
 id                  | integer                  |           | not null | nextval('users_id_seq'::regclass)
 username            | citext                   |           | not null | 
 display_name        | text                     |           |          | 
 avatar_url          | text                     |           |          | 
 created_at          | timestamp with time zone |           | not null | now()
 updated_at          | timestamp with time zone |           | not null | now()
 deleted_at          | timestamp with time zone |           |          | 
 invite_quota        | integer                  |           | not null | 15
 passwd              | text                     |           |          | 
 passwd_reset_code   | text                     |           |          | 
 passwd_reset_time   | timestamp with time zone |           |          | 
 site_admin          | boolean                  |           | not null | false
 page_views          | integer                  |           | not null | 0
 search_queries      | integer                  |           | not null | 0
 tags                | text[]                   |           |          | '{}'::text[]
 billing_customer_id | text                     |           |          | 
Indexes:
    "users_pkey" PRIMARY KEY, btree (id)
    "users_billing_customer_id" UNIQUE, btree (billing_customer_id) WHERE deleted_at IS NULL
    "users_username" UNIQUE, btree (username) WHERE deleted_at IS NULL
Check constraints:
    "users_display_name_max_length" CHECK (char_length(display_name) <= 255)
    "users_username_max_length" CHECK (char_length(username::text) <= 255)
    "users_username_valid_chars" CHECK (username ~ '^[a-zA-Z0-9](?:[a-zA-Z0-9]|-(?=[a-zA-Z0-9]))*$'::citext)
Referenced by:
    TABLE "access_tokens" CONSTRAINT "access_tokens_creator_user_id_fkey" FOREIGN KEY (creator_user_id) REFERENCES users(id)
    TABLE "access_tokens" CONSTRAINT "access_tokens_subject_user_id_fkey" FOREIGN KEY (subject_user_id) REFERENCES users(id)
    TABLE "comments" CONSTRAINT "comments_author_user_id_fkey" FOREIGN KEY (author_user_id) REFERENCES users(id) ON DELETE RESTRICT
    TABLE "discussion_comments" CONSTRAINT "discussion_comments_author_user_id_fkey" FOREIGN KEY (author_user_id) REFERENCES users(id) ON DELETE RESTRICT
    TABLE "discussion_mail_reply_tokens" CONSTRAINT "discussion_mail_reply_tokens_user_id_fkey" FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT
    TABLE "discussion_threads" CONSTRAINT "discussion_threads_author_user_id_fkey" FOREIGN KEY (author_user_id) REFERENCES users(id) ON DELETE RESTRICT
    TABLE "names" CONSTRAINT "names_user_id_fkey" FOREIGN KEY (user_id) REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE
    TABLE "org_invitations" CONSTRAINT "org_invitations_recipient_user_id_fkey" FOREIGN KEY (recipient_user_id) REFERENCES users(id)
    TABLE "org_invitations" CONSTRAINT "org_invitations_sender_user_id_fkey" FOREIGN KEY (sender_user_id) REFERENCES users(id)
    TABLE "org_members" CONSTRAINT "org_members_user_id_fkey" FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT
    TABLE "product_subscriptions" CONSTRAINT "product_subscriptions_user_id_fkey" FOREIGN KEY (user_id) REFERENCES users(id)
    TABLE "registry_extension_releases" CONSTRAINT "registry_extension_releases_creator_user_id_fkey" FOREIGN KEY (creator_user_id) REFERENCES users(id)
    TABLE "registry_extensions" CONSTRAINT "registry_extensions_publisher_user_id_fkey" FOREIGN KEY (publisher_user_id) REFERENCES users(id)
    TABLE "saved_searches" CONSTRAINT "saved_searches_user_id_fkey" FOREIGN KEY (user_id) REFERENCES users(id)
    TABLE "settings" CONSTRAINT "settings_author_user_id_fkey" FOREIGN KEY (author_user_id) REFERENCES users(id) ON DELETE RESTRICT
    TABLE "settings" CONSTRAINT "settings_user_id_fkey" FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT
    TABLE "survey_responses" CONSTRAINT "survey_responses_user_id_fkey" FOREIGN KEY (user_id) REFERENCES users(id)
    TABLE "threads_assignees" CONSTRAINT "threads_assignees_user_id_fkey" FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
    TABLE "threads" CONSTRAINT "threads_creator_user_id_fkey" FOREIGN KEY (creator_user_id) REFERENCES users(id) ON DELETE RESTRICT
    TABLE "user_emails" CONSTRAINT "user_emails_user_id_fkey" FOREIGN KEY (user_id) REFERENCES users(id)
    TABLE "user_external_accounts" CONSTRAINT "user_external_accounts_user_id_fkey" FOREIGN KEY (user_id) REFERENCES users(id)

```
