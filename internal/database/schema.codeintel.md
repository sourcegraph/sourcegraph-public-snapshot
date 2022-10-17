# Table "public.lsif_data_definitions"
```
     Column     |  Type   | Collation | Nullable | Default 
----------------+---------+-----------+----------+---------
 dump_id        | integer |           | not null | 
 scheme         | text    |           | not null | 
 identifier     | text    |           | not null | 
 data           | bytea   |           |          | 
 schema_version | integer |           | not null | 
 num_locations  | integer |           | not null | 
Indexes:
    "lsif_data_definitions_pkey" PRIMARY KEY, btree (dump_id, scheme, identifier)
    "lsif_data_definitions_dump_id_schema_version" btree (dump_id, schema_version)
Triggers:
    lsif_data_definitions_schema_versions_insert AFTER INSERT ON lsif_data_definitions REFERENCING NEW TABLE AS newtab FOR EACH STATEMENT EXECUTE FUNCTION update_lsif_data_definitions_schema_versions_insert()

```

Associates (document, range) pairs with the import monikers attached to the range.

**data**: A gob-encoded payload conforming to an array of [LocationData](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.26/-/blob/enterprise/lib/codeintel/semantic/types.go#L106:6) types.

**dump_id**: The identifier of the associated dump in the lsif_uploads table (state=completed).

**identifier**: The moniker identifier.

**num_locations**: The number of locations stored in the data field.

**schema_version**: The schema version of this row - used to determine presence and encoding of data.

**scheme**: The moniker scheme.

# Table "public.lsif_data_definitions_schema_versions"
```
       Column       |  Type   | Collation | Nullable | Default 
--------------------+---------+-----------+----------+---------
 dump_id            | integer |           | not null | 
 min_schema_version | integer |           |          | 
 max_schema_version | integer |           |          | 
Indexes:
    "lsif_data_definitions_schema_versions_pkey" PRIMARY KEY, btree (dump_id)
    "lsif_data_definitions_schema_versions_dump_id_schema_version_bo" btree (dump_id, min_schema_version, max_schema_version)

```

Tracks the range of schema_versions for each upload in the lsif_data_definitions table.

**dump_id**: The identifier of the associated dump in the lsif_uploads table.

**max_schema_version**: An upper-bound on the `lsif_data_definitions.schema_version` where `lsif_data_definitions.dump_id = dump_id`.

**min_schema_version**: A lower-bound on the `lsif_data_definitions.schema_version` where `lsif_data_definitions.dump_id = dump_id`.

# Table "public.lsif_data_documents"
```
     Column      |  Type   | Collation | Nullable | Default 
-----------------+---------+-----------+----------+---------
 dump_id         | integer |           | not null | 
 path            | text    |           | not null | 
 data            | bytea   |           |          | 
 schema_version  | integer |           | not null | 
 num_diagnostics | integer |           | not null | 
 ranges          | bytea   |           |          | 
 hovers          | bytea   |           |          | 
 monikers        | bytea   |           |          | 
 packages        | bytea   |           |          | 
 diagnostics     | bytea   |           |          | 
Indexes:
    "lsif_data_documents_pkey" PRIMARY KEY, btree (dump_id, path)
    "lsif_data_documents_dump_id_schema_version" btree (dump_id, schema_version)
Triggers:
    lsif_data_documents_schema_versions_insert AFTER INSERT ON lsif_data_documents REFERENCING NEW TABLE AS newtab FOR EACH STATEMENT EXECUTE FUNCTION update_lsif_data_documents_schema_versions_insert()

```

Stores reference, hover text, moniker, and diagnostic data about a particular text document witin a dump.

**data**: A gob-encoded payload conforming to the [DocumentData](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.26/-/blob/enterprise/lib/codeintel/semantic/types.go#L13:6) type. This field is being migrated across ranges, hovers, monikers, packages, and diagnostics columns and will be removed in a future release of Sourcegraph.

**diagnostics**: A gob-encoded payload conforming to the [Diagnostics](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.26/-/blob/enterprise/lib/codeintel/semantic/types.go#L18:2) field of the DocumentDatatype.

**dump_id**: The identifier of the associated dump in the lsif_uploads table (state=completed).

**hovers**: A gob-encoded payload conforming to the [HoversResults](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.26/-/blob/enterprise/lib/codeintel/semantic/types.go#L15:2) field of the DocumentDatatype.

**monikers**: A gob-encoded payload conforming to the [Monikers](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.26/-/blob/enterprise/lib/codeintel/semantic/types.go#L16:2) field of the DocumentDatatype.

**num_diagnostics**: The number of diagnostics stored in the data field.

**packages**: A gob-encoded payload conforming to the [PackageInformation](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.26/-/blob/enterprise/lib/codeintel/semantic/types.go#L17:2) field of the DocumentDatatype.

**path**: The path of the text document relative to the associated dump root.

**ranges**: A gob-encoded payload conforming to the [Ranges](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.26/-/blob/enterprise/lib/codeintel/semantic/types.go#L14:2) field of the DocumentDatatype.

**schema_version**: The schema version of this row - used to determine presence and encoding of data.

# Table "public.lsif_data_documents_schema_versions"
```
       Column       |  Type   | Collation | Nullable | Default 
--------------------+---------+-----------+----------+---------
 dump_id            | integer |           | not null | 
 min_schema_version | integer |           |          | 
 max_schema_version | integer |           |          | 
Indexes:
    "lsif_data_documents_schema_versions_pkey" PRIMARY KEY, btree (dump_id)
    "lsif_data_documents_schema_versions_dump_id_schema_version_boun" btree (dump_id, min_schema_version, max_schema_version)

```

Tracks the range of schema_versions for each upload in the lsif_data_documents table.

**dump_id**: The identifier of the associated dump in the lsif_uploads table.

**max_schema_version**: An upper-bound on the `lsif_data_documents.schema_version` where `lsif_data_documents.dump_id = dump_id`.

**min_schema_version**: A lower-bound on the `lsif_data_documents.schema_version` where `lsif_data_documents.dump_id = dump_id`.

# Table "public.lsif_data_implementations"
```
     Column     |  Type   | Collation | Nullable | Default 
----------------+---------+-----------+----------+---------
 dump_id        | integer |           | not null | 
 scheme         | text    |           | not null | 
 identifier     | text    |           | not null | 
 data           | bytea   |           |          | 
 schema_version | integer |           | not null | 
 num_locations  | integer |           | not null | 
Indexes:
    "lsif_data_implementations_pkey" PRIMARY KEY, btree (dump_id, scheme, identifier)
    "lsif_data_implementations_dump_id_schema_version" btree (dump_id, schema_version)
Triggers:
    lsif_data_implementations_schema_versions_insert AFTER INSERT ON lsif_data_implementations REFERENCING NEW TABLE AS newtab FOR EACH STATEMENT EXECUTE FUNCTION update_lsif_data_implementations_schema_versions_insert()

```

Associates (document, range) pairs with the implementation monikers attached to the range.

**data**: A gob-encoded payload conforming to an array of [LocationData](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.26/-/blob/enterprise/lib/codeintel/semantic/types.go#L106:6) types.

**dump_id**: The identifier of the associated dump in the lsif_uploads table (state=completed).

**identifier**: The moniker identifier.

**num_locations**: The number of locations stored in the data field.

**schema_version**: The schema version of this row - used to determine presence and encoding of data.

**scheme**: The moniker scheme.

# Table "public.lsif_data_implementations_schema_versions"
```
       Column       |  Type   | Collation | Nullable | Default 
--------------------+---------+-----------+----------+---------
 dump_id            | integer |           | not null | 
 min_schema_version | integer |           |          | 
 max_schema_version | integer |           |          | 
Indexes:
    "lsif_data_implementations_schema_versions_pkey" PRIMARY KEY, btree (dump_id)
    "lsif_data_implementations_schema_versions_dump_id_schema_versio" btree (dump_id, min_schema_version, max_schema_version)

```

Tracks the range of schema_versions for each upload in the lsif_data_implementations table.

**dump_id**: The identifier of the associated dump in the lsif_uploads table.

**max_schema_version**: An upper-bound on the `lsif_data_implementations.schema_version` where `lsif_data_implementations.dump_id = dump_id`.

**min_schema_version**: A lower-bound on the `lsif_data_implementations.schema_version` where `lsif_data_implementations.dump_id = dump_id`.

# Table "public.lsif_data_metadata"
```
      Column       |  Type   | Collation | Nullable | Default 
-------------------+---------+-----------+----------+---------
 dump_id           | integer |           | not null | 
 num_result_chunks | integer |           |          | 
Indexes:
    "lsif_data_metadata_pkey" PRIMARY KEY, btree (dump_id)

```

Stores the number of result chunks associated with a dump.

**dump_id**: The identifier of the associated dump in the lsif_uploads table (state=completed).

**num_result_chunks**: A bound of populated indexes in the lsif_data_result_chunks table for the associated dump. This value is used to hash identifiers into the result chunk index to which they belong.

# Table "public.lsif_data_references"
```
     Column     |  Type   | Collation | Nullable | Default 
----------------+---------+-----------+----------+---------
 dump_id        | integer |           | not null | 
 scheme         | text    |           | not null | 
 identifier     | text    |           | not null | 
 data           | bytea   |           |          | 
 schema_version | integer |           | not null | 
 num_locations  | integer |           | not null | 
Indexes:
    "lsif_data_references_pkey" PRIMARY KEY, btree (dump_id, scheme, identifier)
    "lsif_data_references_dump_id_schema_version" btree (dump_id, schema_version)
Triggers:
    lsif_data_references_schema_versions_insert AFTER INSERT ON lsif_data_references REFERENCING NEW TABLE AS newtab FOR EACH STATEMENT EXECUTE FUNCTION update_lsif_data_references_schema_versions_insert()

```

Associates (document, range) pairs with the export monikers attached to the range.

**data**: A gob-encoded payload conforming to an array of [LocationData](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.26/-/blob/enterprise/lib/codeintel/semantic/types.go#L106:6) types.

**dump_id**: The identifier of the associated dump in the lsif_uploads table (state=completed).

**identifier**: The moniker identifier.

**num_locations**: The number of locations stored in the data field.

**schema_version**: The schema version of this row - used to determine presence and encoding of data.

**scheme**: The moniker scheme.

# Table "public.lsif_data_references_schema_versions"
```
       Column       |  Type   | Collation | Nullable | Default 
--------------------+---------+-----------+----------+---------
 dump_id            | integer |           | not null | 
 min_schema_version | integer |           |          | 
 max_schema_version | integer |           |          | 
Indexes:
    "lsif_data_references_schema_versions_pkey" PRIMARY KEY, btree (dump_id)
    "lsif_data_references_schema_versions_dump_id_schema_version_bou" btree (dump_id, min_schema_version, max_schema_version)

```

Tracks the range of schema_versions for each upload in the lsif_data_references table.

**dump_id**: The identifier of the associated dump in the lsif_uploads table.

**max_schema_version**: An upper-bound on the `lsif_data_references.schema_version` where `lsif_data_references.dump_id = dump_id`.

**min_schema_version**: A lower-bound on the `lsif_data_references.schema_version` where `lsif_data_references.dump_id = dump_id`.

# Table "public.lsif_data_result_chunks"
```
 Column  |  Type   | Collation | Nullable | Default 
---------+---------+-----------+----------+---------
 dump_id | integer |           | not null | 
 idx     | integer |           | not null | 
 data    | bytea   |           |          | 
Indexes:
    "lsif_data_result_chunks_pkey" PRIMARY KEY, btree (dump_id, idx)

```

Associates result set identifiers with the (document path, range identifier) pairs that compose the set.

**data**: A gob-encoded payload conforming to the [ResultChunkData](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.26/-/blob/enterprise/lib/codeintel/semantic/types.go#L76:6) type.

**dump_id**: The identifier of the associated dump in the lsif_uploads table (state=completed).

**idx**: The unique result chunk index within the associated dump. Every result set identifier present should hash to this index (modulo lsif_data_metadata.num_result_chunks).

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

# Table "public.rockskip_ancestry"
```
  Column   |         Type          | Collation | Nullable |                    Default                    
-----------+-----------------------+-----------+----------+-----------------------------------------------
 id        | integer               |           | not null | nextval('rockskip_ancestry_id_seq'::regclass)
 repo_id   | integer               |           | not null | 
 commit_id | character varying(40) |           | not null | 
 height    | integer               |           | not null | 
 ancestor  | integer               |           | not null | 
Indexes:
    "rockskip_ancestry_pkey" PRIMARY KEY, btree (id)
    "rockskip_ancestry_repo_id_commit_id_key" UNIQUE CONSTRAINT, btree (repo_id, commit_id)
    "rockskip_ancestry_repo_commit_id" btree (repo_id, commit_id)

```

# Table "public.rockskip_repos"
```
      Column      |           Type           | Collation | Nullable |                  Default                   
------------------+--------------------------+-----------+----------+--------------------------------------------
 id               | integer                  |           | not null | nextval('rockskip_repos_id_seq'::regclass)
 repo             | text                     |           | not null | 
 last_accessed_at | timestamp with time zone |           | not null | 
Indexes:
    "rockskip_repos_pkey" PRIMARY KEY, btree (id)
    "rockskip_repos_repo_key" UNIQUE CONSTRAINT, btree (repo)
    "rockskip_repos_last_accessed_at" btree (last_accessed_at)
    "rockskip_repos_repo" btree (repo)

```

# Table "public.rockskip_symbols"
```
 Column  |   Type    | Collation | Nullable |                   Default                    
---------+-----------+-----------+----------+----------------------------------------------
 id      | integer   |           | not null | nextval('rockskip_symbols_id_seq'::regclass)
 added   | integer[] |           | not null | 
 deleted | integer[] |           | not null | 
 repo_id | integer   |           | not null | 
 path    | text      |           | not null | 
 name    | text      |           | not null | 
Indexes:
    "rockskip_symbols_pkey" PRIMARY KEY, btree (id)
    "rockskip_symbols_gin" gin (singleton_integer(repo_id) gin__int_ops, added gin__int_ops, deleted gin__int_ops, name gin_trgm_ops, singleton(name), singleton(lower(name)), path gin_trgm_ops, singleton(path), path_prefixes(path), singleton(lower(path)), path_prefixes(lower(path)), singleton(get_file_extension(path)), singleton(get_file_extension(lower(path))))
    "rockskip_symbols_repo_id_path_name" btree (repo_id, path, name)

```
