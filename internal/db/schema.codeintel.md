# Table "public.codeintel_schema_migrations"
```
 Column  |  Type   | Modifiers 
---------+---------+-----------
 version | bigint  | not null
 dirty   | boolean | not null
Indexes:
    "codeintel_schema_migrations_pkey" PRIMARY KEY, btree (version)

```

# Table "public.lsif_data_definitions"
```
   Column   |  Type   | Modifiers 
------------+---------+-----------
 dump_id    | integer | not null
 scheme     | text    | not null
 identifier | text    | not null
 data       | bytea   | 
Indexes:
    "lsif_data_definitions_pkey" PRIMARY KEY, btree (dump_id, scheme, identifier)

```

# Table "public.lsif_data_documents"
```
 Column  |  Type   | Modifiers 
---------+---------+-----------
 dump_id | integer | not null
 path    | text    | not null
 data    | bytea   | 
Indexes:
    "lsif_data_documents_pkey" PRIMARY KEY, btree (dump_id, path)

```

# Table "public.lsif_data_metadata"
```
      Column       |  Type   | Modifiers 
-------------------+---------+-----------
 dump_id           | integer | not null
 num_result_chunks | integer | 
Indexes:
    "lsif_data_metadata_pkey" PRIMARY KEY, btree (dump_id)

```

# Table "public.lsif_data_references"
```
   Column   |  Type   | Modifiers 
------------+---------+-----------
 dump_id    | integer | not null
 scheme     | text    | not null
 identifier | text    | not null
 data       | bytea   | 
Indexes:
    "lsif_data_references_pkey" PRIMARY KEY, btree (dump_id, scheme, identifier)

```

# Table "public.lsif_data_result_chunks"
```
 Column  |  Type   | Modifiers 
---------+---------+-----------
 dump_id | integer | not null
 idx     | integer | not null
 data    | bytea   | 
Indexes:
    "lsif_data_result_chunks_pkey" PRIMARY KEY, btree (dump_id, idx)

```
