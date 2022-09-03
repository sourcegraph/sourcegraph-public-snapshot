# Table "public.console_users"
```
 Column |  Type   | Collation | Nullable |                  Default                  
--------+---------+-----------+----------+-------------------------------------------
 id     | integer |           | not null | nextval('console_users_id_seq'::regclass)
 email  | citext  |           | not null | 
Indexes:
    "console_users_pkey" PRIMARY KEY, btree (id)

```

# Table "public.migration_logs"
```
            Column             |           Type           | Collation | Nullable |                  Default                   
-------------------------------+--------------------------+-----------+----------+--------------------------------------------
 id                            | integer                  |           | not null | nextval('migration_logs_id_seq'::regclass)
 migration_logs_schema_version | integer                  |           | not null | 
 schema                        | text                     |           | not null | 
 version                       | integer                  |           | not null | 
 up                            | boolean                  |           | not null | 
 started_at                    | timestamp with time zone |           | not null | 
 finished_at                   | timestamp with time zone |           |          | 
 success                       | boolean                  |           |          | 
 error_message                 | text                     |           |          | 
 backfilled                    | boolean                  |           | not null | false
Indexes:
    "migration_logs_pkey" PRIMARY KEY, btree (id)

```
